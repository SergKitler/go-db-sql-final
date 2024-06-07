package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	p, err := store.Add(parcel)
	if err != nil {
		require.NoError(t, err)
		return
	}
	assert.NotEmpty(t, p)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	parClient, err := store.Get(p)
	if err != nil {
		require.NoError(t, err)
	}
	assert.Equal(t, parcel.Client, parClient.Client)
	assert.Equal(t, parcel.Status, parClient.Status)
	assert.Equal(t, parcel.Address, parClient.Address)
	assert.Equal(t, parcel.CreatedAt, parClient.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(p)
	if err != nil {
		require.NoError(t, err)
	}

	parClient, err = store.Get(p)
	if err != nil {
		require.Equal(t, sql.ErrNoRows, err)
	}

	assert.Empty(t, parClient.Client)
	assert.Empty(t, parClient.Status)
	assert.Empty(t, parClient.Address)
	assert.Empty(t, parClient.CreatedAt)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	p, err := store.Add(parcel)
	if err != nil {
		require.NoError(t, err)
		return
	}
	assert.NotEmpty(t, p)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(p, newAddress)
	if err != nil {
		require.NoError(t, err)
		return
	}

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	var parClient Parcel
	parClient, err = store.Get(p)
	if err != nil {
		require.NoError(t, err)
	}
	assert.Equal(t, parcel.Client, parClient.Client)
	assert.Equal(t, parcel.Status, parClient.Status)
	assert.Equal(t, newAddress, parClient.Address)
	assert.Equal(t, parcel.CreatedAt, parClient.CreatedAt)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	p, err := store.Add(parcel)
	if err != nil {
		require.NoError(t, err)
		return
	}
	assert.NotEmpty(t, p)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := ParcelStatusSent

	err = store.SetStatus(p, newStatus)
	if err != nil {
		require.NoError(t, err)
		return
	}

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	var parClient Parcel
	parClient, err = store.Get(p)
	if err != nil {
		require.NoError(t, err)
	}
	assert.Equal(t, parcel.Client, parClient.Client)
	assert.Equal(t, newStatus, parClient.Status)
	assert.Equal(t, parcel.Address, parClient.Address)
	assert.Equal(t, parcel.CreatedAt, parClient.CreatedAt)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		if err != nil {
			require.NoError(t, err)
			return
		}
		assert.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		require.NoError(t, err)
		return
	}
	assert.Equal(t, len(parcels), len(storedParcels))

	// check
	// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
	// убедитесь, что все посылки из storedParcels есть в parcelMap
	// убедитесь, что значения полей полученных посылок заполнены верно

	for _, parcel := range storedParcels {
		id := parcel.Number
		assert.Equal(t, parcelMap[id], parcel)
		assert.Equal(t, parcelMap[id].Client, parcel.Client)
		assert.Equal(t, parcelMap[id].Address, parcel.Address)
		assert.Equal(t, parcelMap[id].Status, parcel.Status)
		assert.Equal(t, parcelMap[id].CreatedAt, parcel.CreatedAt)
	}
}
