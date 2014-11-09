package main

//TODO: probably name this one 'archiver' and rename 'archiver.go' to 'giles.go'

func AddData(readings []interface{}) bool {
	return true
}

func GetData(streamids []string, start, end uint64) ([]interface{}, error) {
	return []interface{}{}, nil
}

func PrevData(streamids []string, start uint64, limit int32) ([]interface{}, error) {
	return []interface{}{}, nil
}

func NextData(streamids []string, start uint64, limit int32) ([]interface{}, error) {
	return []interface{}{}, nil
}

func GetTags(select_tags, where_tags map[string]interface{}) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

func GetUUIDs(where_tags map[string]interface{}) ([]string, error) {
	return []string{}, nil
}

func SetTags(update_tags, where_tags map[string]interface{}) (int, error) {
	return 0, nil
}
