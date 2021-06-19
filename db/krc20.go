// Package db
package db

const (
	cKRC20 = "KRC20"
)

type IKrc20 interface {
}

type KRC20Filter struct {
	Address string
	Name    string
	Symbol  string
}

//
//func (m *mongoDB) KRC20(ctx context.Context, filter KRC20Filter) (*types.KRC20, error) {
//	var krc20 *types.KRC20
//	if err := m.wrapper.C(cKRC20).FindOne().Decode(&krc20); err != nil {
//		return nil, err
//	}
//	return krc20, nil
//}

func (m *mongoDB) InsertKRC20() {

}

func (m *mongoDB) UpdateKRC20() {

}
