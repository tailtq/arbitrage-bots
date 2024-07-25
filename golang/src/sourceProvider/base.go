package sourceProvider

type SourceProvider interface {
    GetTokenList() (*[]Symbol, error)
}

type Symbol struct {
    Symbol string `json:"symbol"`
    BaseAsset string `json:"baseAsset"`
    QuoteAsset string `json:"quoteAsset"`
}
