package nsfContext

import (
	"encoding/json"
	"fmt"
	"nsf/internal/db"
	"nsf/logger"
)

type serviceInfo struct {
	NfName     string `json:"nf_name"`
	SbiAddress string `json:"sbi_address"`
	SbiPort    int    `json:"sbi_port"`
}

// TODO: You may need to add some fields
type serviceContextIE struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: You may need to add some fields
type serviceContext struct {
	db.DbIf

	*logger.NsfLogger
}

// TODO: Finish the initialization of the serviceContext instance
func newServiceContext(serviceContextIE *serviceContextIE) *serviceContext {

	return &serviceContext{
		DbIf: serviceContextIE.DbIf,

		NsfLogger: serviceContextIE.NsfLogger,
	}
}

// TODO: Finish the release of the serviceContext instance
func (s *serviceContext) release() {
	s.NsfLogger.SvcLog.Infoln("Release serviceContext...")

	s.NsfLogger.SvcLog.Infoln("serviceContext released")
}

// TODO: Finish the implementation of the registerServices method with an error return type
func (s *serviceContext) registerServices(nfName, sbiAddr string, sbiPort int, services []string) error {
	s.NsfLogger.SvcLog.Tracef("Registering %s services", nfName)
	info := serviceInfo{
		NfName:     nfName,
		SbiAddress: sbiAddr,
		SbiPort:    sbiPort,
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	records := make(map[string]string)

	for _, service := range services {
		records[service] = string(data)
	}

	return s.DbIf.SaveAll(records)
}

// TODO: Finish the implementation of the deregisterServices method with an error return type
func (s *serviceContext) deregisterServices(nfName string, services []string) error {
	s.NsfLogger.SvcLog.Tracef("De-registering %s services", nfName)

	for _, service := range services {
		// 查詢這項服務的提供者
		owner, _, _, err := s.discoverService(service)
		if err != nil {
			return err
		}

		// 確認服務屬於要求刪除的 NF
		if owner != nfName {
			return fmt.Errorf(
				"service %s belongs to %s, not %s",
				service, owner, nfName,
			)
		}

		// 從資料庫移除這項服務
		err = s.DbIf.Remove(service)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Finish the implementation of the discoverService method with these return values:
// (string, string, int, error) representing the NF name, SBI address, SBI port, and an error if any
func (s *serviceContext) discoverService(service string) (string, string, int, error) {
	s.NsfLogger.SvcLog.Tracef("Discovering service %s", service)

	value, err := s.DbIf.Load(service)
	if err != nil {
		return "", "", 0, err
	}
	// 空字串表示沒有我們儲存的服務紀錄
	if value == "" {
		return "", "", 0, fmt.Errorf("service %s not found", service)
	}

	// 準備接收解析結果
	var info serviceInfo

	// 將 JSON 還原成 struct
	err = json.Unmarshal([]byte(value), &info)
	if err != nil {
		return "", "", 0, err
	}

	// 回傳查到的服務資訊
	return info.NfName, info.SbiAddress, info.SbiPort, nil
}
