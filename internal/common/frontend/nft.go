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
	NftAmount      null.Int           `json:"nft_amount"`
	TotalNftAmount null.Int           `json:"total_nft_amount"`
	Likes          int                `json:"likes"`
	IsLiked        bool               `json:"is_liked"`
	ContentUrl     string             `json:"content_url"`
	ImageUrl       string             `json:"image_url"`
	ContentType    NftContentType     `json:"content_type"`
	Categories     []string           `json:"categories"`
	CollectionId   null.Int           `json:"collection_id"`
	Collection     NftCollectionModel `json:"collection"`
	ContractId     string             `json:"contract_id"`
	TokenId        string             `json:"token_id"`
	IsOwned        bool               `json:"is_owned"`
	Listings       []NftListingModel  `json:"listings"`
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

type NftListingModel struct {
	OnSale           []NftContentItemGroup `json:"on_sale"`
	NotForSaleAmount int                   `json:"not_for_sale_amount"`
	TotalAmount      int                   `json:"total_amount"`
	CurrentOwnerId   int64                 `json:"current_owner_id"`
	CurrentOwner     NftShortUserModel     `json:"current_owner"`
}

type NftContentItemGroup struct {
	Price  NftPriceModel `json:"price"`
	Amount int           `json:"amount"`
}

type NftPriceModel struct {
	Price    decimal.Decimal `json:"price"`
	Currency common.Currency `json:"currency"`
	UsdPrice decimal.Decimal `json:"usd_price"`
}

type NftCollectionModel struct {
	Id             int               `json:"id"`
	ImageUrl       string            `json:"image_url"`
	Type           NftCollectionType `json:"type"`
	SupplyLazyMint bool              `json:"supply_lazy_mint"`
	Name           string            `json:"name"`
	Hash           string            `json:"hash"`
}
