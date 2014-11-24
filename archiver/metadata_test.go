package archiver

import (
	"code.google.com/p/go-uuid/uuid"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func BenchmarkSaveMetadata(b *testing.B) {
	store = NewStore(*mongoip, *mongoport)
	myuuid := uuid.New()
	sm := &SmapMessage{
		Metadata: bson.M{"X": 3, "Y": 4, "Z": 5},
		UUID:     myuuid,
		Path:     "/test/sensor",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.SaveMetadata(sm)
		//store.TagsUUID(myuuid)
	}
}

func BenchmarkTagsUUID(b *testing.B) {
	store = NewStore(*mongoip, *mongoport)
	myuuid := uuid.New()
	sm := &SmapMessage{
		Metadata: bson.M{"X": 3, "Y": 4, "Z": 5},
		UUID:     myuuid,
		Path:     "/test/sensor",
	}
	store.SaveMetadata(sm)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.TagsUUID(myuuid)
	}
}

func TestSaveMetadata(t *testing.T) {
	store = NewStore(*mongoip, *mongoport)
	myuuid := uuid.New()
	sm := &SmapMessage{
		Metadata: bson.M{"X": 3, "Y": 4, "Z": 5},
		UUID:     myuuid,
		Path:     "/test/sensor",
	}
	store.SaveMetadata(sm)
	jsonbytes, err := store.TagsUUID(myuuid)
	if err != nil {
		t.Error(err)
	}
	if len(jsonbytes) == 0 {
		t.Error("No data returned for UUID", myuuid)
	}
}
