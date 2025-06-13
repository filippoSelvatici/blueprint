package workloadgen

import (
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
	"github.com/google/uuid"
)

func RandomSessionID() string {
	return uuid.NewString()
}

func GenRunRegister() (string, string) {
	return uuid.NewString(), uuid.NewString()
}

func GenPostCard() user.Card {
	return user.Card{
		LongNum: uuid.NewString(),
		Expires: "2035-04-30",
		CCV:     "000",
		ID:      uuid.NewString(),
	}
}

func GenPostAddress() user.Address {
	return user.Address{
		Street:   "Via Roma",
		Number:   "23",
		Country:  "Switzerland",
		City:     "Zurich",
		PostCode: "2929",
		ID:       uuid.NewString(),
	}
}
