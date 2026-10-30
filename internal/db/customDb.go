package db

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type customDb struct {
	mu   sync.RWMutex
	path string
	data map[string]string
}

func newCustomDb(path string) (*customDb, error) {
	// 建立物件，先準備一個空 map
	d := &customDb{
		path: path,
		data: make(map[string]string),
	}

	// 讀取之前保存的檔案
	content, err := os.ReadFile(path)

	// 第一次使用，檔案不存在是正常情況
	if os.IsNotExist(err) {
		return d, nil
	}

	// 其他讀取錯誤，例如沒有權限
	if err != nil {
		return nil, fmt.Errorf("read custom DB: %w", err)
	}

	// 把檔案中的 JSON 還原到 map
	if err := json.Unmarshal(content, &d.data); err != nil {
		return nil, fmt.Errorf("decode custom DB: %w", err)
	}

	// JSON 若是 null，解析後 map 會變成 nil，需要重新初始化
	if d.data == nil {
		d.data = make(map[string]string)
	}

	return d, nil
}

// 呼叫此方法前，必須持有 d.mu 的寫鎖。
func (d *customDb) persistLocked(data map[string]string) error {
	content, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode custom DB: %w", err)
	}

	tempPath := d.path + ".tmp"

	if err := os.WriteFile(tempPath, content, 0600); err != nil {
		return fmt.Errorf("write custom DB temp file: %w", err)
	}

	if err := os.Rename(tempPath, d.path); err != nil {
		return fmt.Errorf("replace custom DB file: %w", err)
	}

	return nil
}

func (d *customDb) Save(key, value string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 複製目前資料，準備修改
	next := make(map[string]string, len(d.data)+1)
	for k, v := range d.data {
		next[k] = v
	}

	// 在副本新增或更新資料
	next[key] = value

	// 先確認寫入檔案成功
	if err := d.persistLocked(next); err != nil {
		return err
	}

	// 成功後，才更新記憶體中的資料
	d.data = next
	return nil
}

func (d *customDb) SaveAll(data map[string]string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 複製目前資料
	next := make(map[string]string, len(d.data)+len(data))
	for k, v := range d.data {
		next[k] = v
	}

	// 將這次傳入的資料合併到副本
	for k, v := range data {
		next[k] = v
	}

	// 整批寫入成功後，再更新記憶體
	if err := d.persistLocked(next); err != nil {
		return err
	}

	d.data = next
	return nil
}

func (d *customDb) Load(key string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	value, ok := d.data[key]
	if !ok {
		return "", nil
	}

	return value, nil
}

func (d *customDb) Exist(key string) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, ok := d.data[key]
	return ok, nil
}

func (d *customDb) LoadAll() (map[string]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]string, len(d.data))
	for k, v := range d.data {
		result[k] = v
	}

	return result, nil
}

func (d *customDb) Remove(key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 複製目前資料
	next := make(map[string]string, len(d.data))
	for k, v := range d.data {
		next[k] = v
	}

	// 從副本移除指定的 key
	delete(next, key)

	// 寫檔成功後，才更新記憶體
	if err := d.persistLocked(next); err != nil {
		return err
	}

	d.data = next
	return nil
}

func (d *customDb) RemoveAll() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 準備空 map
	next := make(map[string]string)

	// 將空資料寫入檔案
	if err := d.persistLocked(next); err != nil {
		return err
	}

	// 寫檔成功後，再清空記憶體中的資料
	d.data = next
	return nil
}

func (d *customDb) Release() error {
	return nil
}
