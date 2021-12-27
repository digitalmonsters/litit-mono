package frontend

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
)

type NftContentType byte

const (
	None  NftContentType = 0
	Photo NftContentType = 1
	Video NftContentType = 2
	Audio NftContentType = 3
)

type NftCollectionType string

const (
	CollectionTypeERC721  = NftCollectionType("ERC721")
	CollectionTypeERC1155 = NftCollectionType("ERC1155")
)

type NftContentModel struct {
	Id             int64              `json:"id"`
	Name           string             `json:"name"`
	AuthorId       int64              `json:"author_id"`
	Author         NftShortUserModel  `json:"author"`
	Description    string             `json:"description"`
	CurrentBid     NftBidModel        `json:"current_bid"`
	EndDate        null.Time          `json:"end_date"`
	NftAmount      null.Int           `json:"nft_amount"`
	TotalNftAmount null.Int           `json:"total_nft_amount"`
	CanBid         bool               `json:"can_bid"`
	Likes          int                `json:"likes"`
	IsLiked        bool               `json:"is_liked"`
	ContentUrl     string             `json:"content_url"`
	ImageUrl       string             `json:"image_url"`
	ContentType    NftContentType     `json:"content_type"`
	Categories     []string           `json:"categories"`
	CollectionId   null.Int           `json:"collection_id"`
	Collection     NftCollectionModel `json:"collection"`
	Hash           string             `json:"hash"`
	Price          NftPriceModel      `json:"price"`
	IsOwned        bool               `json:"is_owned"`
	IsOnSale       bool               `json:"is_on_sale"`
}

type NftShortUserModel struct {
	Id            int64       `json:"id"`
	FirstName     null.String `json:"first_name"`
	UserName      string      `json:"user_name"`
	Avatar        string      `json:"avatar"`
	Online        bool        `json:"online"`
	IsFollowed    bool        `json:"is_followed"`
	WalletAddress string      `json:"wallet_address"`
}

type NftPriceModel struct {
	Price    decimal.Decimal `json:"price"`
	Currency common.Currency `json:"currency"`
	UsdPrice decimal.Decimal `json:"usd_price"`
}

type NftBidModel struct {
	Bid         decimal.Decimal `json:"bid"`
	BidCurrency common.Currency `json:"bid_currency"`
	UsdPrice    decimal.Decimal `json:"usd_price"`
}

type NftCollectionModel struct {
	Id             int               `json:"id"`
	ImageUrl       string            `json:"image_url"`
	Type           NftCollectionType `json:"type"`
	SupplyLazyMint bool              `json:"supply_lazy_mint"`
	Name           string            `json:"name"`
	Hash           string            `json:"hash"`
}
