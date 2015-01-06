package mphandler

import (
	"testing"
)

/*
[134 168 77 101 116 97 100 97 116 97 128 164 80 97 116 104 169 47 115 101 110
115 111 114 47 48 170 80 114 111 112 101 114 116 105 101 115 128 168 82 101 97
100 105 110 103 115 130 168 82 101 97 100 105 110 103 115 145 146 206 5 93 74
128 0 164 117 117 105 100 218 0 36 97 99 99 98 48 101 100 101 45 57 52 57 98 45
49 49 101 52 45 56 54 56 50 45 54 48 48 51 48 56 57 101 100 49 100 48 163 107
101 121 218 0 88 112 105 104 108 72 97 85 89 81 71 99 103 79 108 101 79 45 108
53 45 102 103 54 45 87 120 121 80 74 119 55 54 115 52 111 114 99 114 112 65 48
74 67 95 118 56 114 49 119 120 90 105 87 117 49 79 68 104 107 108 76 119 99 115
57 66 65 88 115 54 66 48 83 111 97 103 103 100 51 109 70 99 74 89 86 119 61 61
164 117 117 105 100 218 0 36 97 99 99 98 48 101 100 101 45 57 52 57 98 45 49 49
101 52 45 56 54 56 50 45 54 48 48 51 48 56 57 101 100 49 100 48]
*/

var mybytes = []byte{134, 168, 77, 101, 116, 97, 100, 97, 116, 97, 128, 164, 80, 97, 116, 104, 169, 47, 115, 101, 110, 115, 111, 114, 47, 48, 170, 80, 114, 111, 112, 101, 114, 116, 105, 101, 115, 128, 168, 82, 101, 97, 100, 105, 110, 103, 115, 130, 168, 82, 101, 97, 100, 105, 110, 103, 115, 145, 146, 206, 5, 93, 74, 128, 0, 164, 117, 117, 105, 100, 218, 0, 36, 97, 99, 99, 98, 48, 101, 100, 101, 45, 57, 52, 57, 98, 45, 49, 49, 101, 52, 45, 56, 54, 56, 50, 45, 54, 48, 48, 51, 48, 56, 57, 101, 100, 49, 100, 48, 163, 107, 101, 121, 218, 0, 88, 112, 105, 104, 108, 72, 97, 85, 89, 81, 71, 99, 103, 79, 108, 101, 79, 45, 108, 53, 45, 102, 103, 54, 45, 87, 120, 121, 80, 74, 119, 55, 54, 115, 52, 111, 114, 99, 114, 112, 65, 48, 74, 67, 95, 118, 56, 114, 49, 119, 120, 90, 105, 87, 117, 49, 79, 68, 104, 107, 108, 76, 119, 99, 115, 57, 66, 65, 88, 115, 54, 66, 48, 83, 111, 97, 103, 103, 100, 51, 109, 70, 99, 74, 89, 86, 119, 61, 61, 164, 117, 117, 105, 100, 218, 0, 36, 97, 99, 99, 98, 48, 101, 100, 101, 45, 57, 52, 57, 98, 45, 49, 49, 101, 52, 45, 56, 54, 56, 50, 45, 54, 48, 48, 51, 48, 56, 57, 101, 100, 49, 100, 48}

func TestDecode(t *testing.T) {
	decode(mybytes)
}

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		decode(mybytes)
	}
}
