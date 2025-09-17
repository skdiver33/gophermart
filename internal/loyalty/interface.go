package loyalty

type AccrualClientInterface interface {
	GetAccural(number string) (*AccrualForOrder, error)
}
