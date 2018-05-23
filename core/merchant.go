package core

import (
	"fmt"
	"sync"
)

var merchantMaps *sync.Map

func init() {
	merchantMaps = new(sync.Map)
}

type Merchant struct {
	MerchantNo     string `gorm:"not null;type:varchar(80);unique" json:"merchant_no" yaml:"merchant_no"` //商户ID
	MerchantSecret string `gorm:"type:varchar(80)" json:"merchant_secret" yaml:"merchant_secret"`         //商户开发密钥
	MerchantName   string `gorm:"type:varchar(100)" json:"merchant_name" yaml:"merchant_name"`            //商户名称
	IsEnabled      bool   `gorm:"default:false" json:"is_enabled" yaml:"is_enabled"`                      //是否可用
}

// 获取商户配置
func LoadMerchant(key string) (*Merchant, error) {
	iter, ok := merchantMaps.Load(key)
	if ok {
		switch iter.(type) {
		case Merchant:
			m := iter.(Merchant)
			return &m, nil
		default:
			return nil, fmt.Errorf("Merchant don't enabled")
		}
	} else {
		return nil, fmt.Errorf("Merchant don't enabled")
	}
}

// 启用商户
func EnableMerchant(m Merchant) {
	merchantMaps.Store(m.MerchantNo, m)
}

// 停用商户
func DisableMerchant(iter interface{}) {
	switch iter.(type) {
	case *Merchant:
		merchantMaps.Delete(iter.(*Merchant).MerchantNo)
	case string:
		merchantMaps.Delete(iter.(string))
	}
}
