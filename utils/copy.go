package utils

import "encoding/xml"

// 通过xml解析的结构体，des必须是引用
func Copy(src, des interface{})bool{
	bs,err:=xml.Marshal(src)
	if err!=nil {
		return false
	}
	err = xml.Unmarshal(bs,des)
	if err!=nil {
		return false
	}
	return true
}