package service

import (
	"encoding/binary"
	"testing"
)

func TestSimpleCodec_Encode(t *testing.T) {
	tests := []struct {
		name           string
		codec          SimpleCodec
		expectedHeader []byte
		expectedData   []byte
		expectError    bool
	}{
		{
			name: "valid encode",
			codec: SimpleCodec{
				CurrentOrganize: 1,
				CommandCode:     2,
				Data:            []byte("test data"),
			},
			expectedHeader: func() []byte {
				header := make([]byte, headerSize)
				binary.BigEndian.PutUint16(header[0:2], 1)
				binary.BigEndian.PutUint16(header[2:4], 2)
				binary.BigEndian.PutUint32(header[4:8], uint32(len([]byte("test data"))))
				return header
			}(),
			expectedData: []byte("test data"),
			expectError:  false,
		},
		{
			name: "empty data",
			codec: SimpleCodec{
				CurrentOrganize: 1,
				CommandCode:     2,
				Data:            []byte(""),
			},
			expectedHeader: func() []byte {
				header := make([]byte, headerSize)
				binary.BigEndian.PutUint16(header[0:2], 1)
				binary.BigEndian.PutUint16(header[2:4], 2)
				binary.BigEndian.PutUint32(header[4:8], 0)
				return header
			}(),
			expectedData: []byte(""),
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := tt.codec.Encode()
			if (err != nil) != tt.expectError {
				t.Errorf("Encode() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				expected := append(tt.expectedHeader, tt.expectedData...)
				if len(encoded) != len(expected) {
					t.Errorf("Encode() length = %v, expected %v", len(encoded), len(expected))
				}

				for i := range expected {
					if encoded[i] != expected[i] {
						t.Errorf("Encode() byte at %d = %v, expected %v", i, encoded[i], expected[i])
					}
				}
			}
		})
	}
}
