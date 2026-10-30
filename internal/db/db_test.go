package db_test

import (
	"fmt"
	"nsf/internal/db"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	db.NSF_DB_PATH = "/tmp/nsf_db_test.db"

	exitCode := m.Run()

	// 清除 bbolt 測試檔案
	if err := os.Remove(db.NSF_DB_PATH); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to remove test DB file: %v\n", err)
		exitCode = 1
	}

	// 清除自訂 JSON 資料庫的測試檔案
	customPath := db.NSF_DB_PATH + ".json"
	if err := os.Remove(customPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Failed to remove custom test DB file: %v\n", err)
		exitCode = 1
	}

	os.Exit(exitCode)
}

var testDbCases = []struct {
	name   string
	key    string
	valuse string
}{
	{
		name:   "Data 1",
		key:    "key1",
		valuse: "value1",
	},
	{
		name:   "Data 2",
		key:    "key2",
		valuse: "value2",
	},
	{
		name:   "Data 3",
		key:    "key3",
		valuse: "value3",
	},
}

var dbInstance db.DbIf

func TestDb(t *testing.T) {
	t.Run("Bbolt DB", testBboltDb)
	t.Run("Custom DB", testCustomDb)
}

func testBboltDb(t *testing.T) {
	var err error
	dbInstance, err = db.NewDb("bbolt")
	if err != nil {
		t.Fatalf("Failed to create DB instance: %v", err)
	}
	defer func(t *testing.T) {
		if err := dbInstance.Release(); err != nil {
			t.Errorf("Failed to release DB instance: %v", err)
		}
	}(t)

	t.Run("Single Data Operations", testSingleDataOperations)
	t.Run("Multiple Data Operations", testMultipleDataOperations)
}

func testCustomDb(t *testing.T) {
	var err error
	dbInstance, err = db.NewDb("custom") // TODO: change "custom" to the actual name of your custom DB implementation
	if err != nil {
		t.Fatalf("Failed to create DB instance: %v", err)
	}
	defer func(t *testing.T) {
		if err := dbInstance.Release(); err != nil {
			t.Errorf("Failed to release DB instance: %v", err)
		}
	}(t)

	t.Run("Single Data Operations", testSingleDataOperations)
	t.Run("Multiple Data Operations", testMultipleDataOperations)
}

func testSingleDataOperations(t *testing.T) {
	for _, tc := range testDbCases {
		if err := dbInstance.Save(tc.key, tc.valuse); err != nil {
			t.Errorf("Failed to save data for %s: %v", tc.name, err)
		}

		exist, err := dbInstance.Exist(tc.key)
		if err != nil {
			t.Errorf("Failed to check existence for %s: %v", tc.name, err)
		} else if !exist {
			t.Errorf("Data for %s does not exist after saving", tc.name)
		}
	}

	for _, tc := range testDbCases {
		value, err := dbInstance.Load(tc.key)
		if err != nil {
			t.Errorf("Failed to load data for %s: %v", tc.name, err)
		} else if value != tc.valuse {
			t.Errorf("Loaded value for %s is incorrect: got %s, want %s", tc.name, value, tc.valuse)
		}
	}

	for _, tc := range testDbCases {
		if err := dbInstance.Remove(tc.key); err != nil {
			t.Errorf("Failed to delete data for %s: %v", tc.name, err)
		}

		exist, err := dbInstance.Exist(tc.key)
		if err != nil {
			t.Errorf("Failed to check existence after deletion for %s: %v", tc.name, err)
		} else if exist {
			t.Errorf("Data for %s still exists after deletion", tc.name)
		}
	}
}

func testMultipleDataOperations(t *testing.T) {
	dataMap := make(map[string]string)
	for _, tc := range testDbCases {
		dataMap[tc.key] = tc.valuse
	}

	if err := dbInstance.SaveAll(dataMap); err != nil {
		t.Errorf("Failed to save multiple data: %v", err)
	}

	for _, tc := range testDbCases {
		exist, err := dbInstance.Exist(tc.key)
		if err != nil {
			t.Errorf("Failed to check existence for %s: %v", tc.name, err)
		} else if !exist {
			t.Errorf("Data for %s does not exist after saving multiple data", tc.name)
		}
	}

	allData, err := dbInstance.LoadAll()
	if err != nil {
		t.Errorf("Failed to load all data: %v", err)
	} else {
		for _, tc := range testDbCases {
			value, ok := allData[tc.key]
			if !ok {
				t.Errorf("Key %s not found in loaded data", tc.name)
			} else if value != tc.valuse {
				t.Errorf("Loaded value for %s is incorrect: got %s, want %s", tc.name, value, tc.valuse)
			}
		}
	}

	if err := dbInstance.RemoveAll(); err != nil {
		t.Errorf("Failed to remove all data: %v", err)
	}

	for _, tc := range testDbCases {
		exist, err := dbInstance.Exist(tc.key)
		if err != nil {
			t.Errorf("Failed to check existence after removing all data for %s: %v", tc.name, err)
		} else if exist {
			t.Errorf("Data for %s still exists after removing all data", tc.name)
		}
	}
}

func TestCustomDbPersistence(t *testing.T) {
	originalPath := db.NSF_DB_PATH
	db.NSF_DB_PATH = t.TempDir() + "/persistence.db"
	defer func() {
		db.NSF_DB_PATH = originalPath
	}()

	// 第一次開啟，寫入資料。
	first, err := db.NewDb("custom")
	if err != nil {
		t.Fatal(err)
	}

	if err := first.Save("service1", "NF1"); err != nil {
		t.Fatal(err)
	}

	if err := first.Release(); err != nil {
		t.Fatal(err)
	}

	// 建立另一個資料庫物件，從同一個檔案讀取。
	second, err := db.NewDb("custom")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := second.Release(); err != nil {
			t.Error(err)
		}
	}()

	value, err := second.Load("service1")
	if err != nil {
		t.Fatal(err)
	}

	if value != "NF1" {
		t.Fatalf("expected NF1, got %q", value)
	}
}
