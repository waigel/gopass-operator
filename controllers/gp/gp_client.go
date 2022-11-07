package gp

import (
	"context"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/api"
)

func GetGopassSecret(ctx context.Context, secretPath string) (gopass.Secret, error) {
	gp, err := api.New(ctx)
	if err != nil {
		panic(err)
	}
	sec, err := gp.Get(ctx, secretPath, "latest")
	//	fmt.Printf("content of %s: %s\n", secretPath, string(sec.Body()))
	return sec, err
}
