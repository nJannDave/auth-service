package entity

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

type errs struct {
	isErr bool
	message string
}

type generalInstructions struct {
	name string
	fn func() error
	err errs
}

func TestUserData(t *testing.T) {
	const tableName = "account"
	realData := UserData {
		NIK: "1234567891123456",
		Name: "bainn",
		Password: "bain091221",
	}
	data := NewUserData(realData.NIK, realData.Name, realData.Password)
	instr := []generalInstructions {
		{
			name: "success - validate nik",
			fn: func() error {
				return realData.ValidateNik()
			},
			err: errs {
				isErr: false,
				message: "",
			},
		},
		{
			name: "failed - validate nik",
			fn: func() error {
				realData.NIK = "111"
				return realData.ValidateNik()
			},
			err: errs {
				isErr: true,
				message: "nik must be 16 digits",
			},
		},
	}

	require.Equal(t, tableName, realData.TableName())
	require.Equal(t, &realData, data)

	for _, tc := range instr {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.isErr {
				err := tc.fn()
				require.Error(t, err)
			} else {
				err := tc.fn()
				require.NoError(t, err)
			}
		})
	}
}

func TestResidence(t *testing.T) {
	realData := Residence {
		Province: "Jawa Barat",
		City: "Bandung",
	}
	data := NewResidence(realData.Province, realData.City)
	require.Equal(t, &realData, data)
}

func TestJunctionData(t *testing.T) {
	realData := JunctionData {
		ProvinceId: 2,
		AccountId: 10,
		CityId: 3,
	}
	data := NewJuctionData(realData.ProvinceId, realData.AccountId, realData.CityId)
	require.Equal(t, &realData, data)
	require.Equal(t, "junction", realData.TableName())
}