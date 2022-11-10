package eventhandler

import "encoding/json"

func unmarshalObjValue(object interface{}, instance interface{}) error {
	jsonData, err := json.Marshal(object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, instance)
	if err != nil {
		return err
	}
	return nil
}
