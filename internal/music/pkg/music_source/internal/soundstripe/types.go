package soundstripe

type soundstripeSongsResp struct {
	MusicData  []musicData    `json:"data"`
	Pagination pagination     `json:"links"`
	Included   []includedData `json:"included"`
}

type soundstripeSingleSongsResp struct {
	MusicData  musicData      `json:"data"`
	Pagination pagination     `json:"links"`
	Included   []includedData `json:"included"`
}

type musicData struct {
	Id            string     `json:"id"`
	Type          string     `json:"type"`
	Attributes    attributes `json:"attributes"`
	Relationships struct {
		Artists struct {
			Data []struct {
				Id   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"artists"`
		AudioFiles struct {
			Data []struct {
				Id   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"audio_files"`
	} `json:"relationships"`
}

type T struct {
	Relationships struct {
		Artists struct {
			Data []struct {
				Id   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"artists"`
		AudioFiles struct {
			Data []struct {
				Id   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"audio_files"`
	} `json:"relationships"`
}

type attributes struct {
	Title string `json:"title"`
	Tags  tags   `json:"tags"`
}

type tags struct {
	Genre []string `json:"genre"`
}

type pagination struct {
	Meta paginationMeta `json:"meta"`
}

type paginationMeta struct {
	TotalCount int64 `json:"total_count"`
}

type includedData struct {
	Id         string             `json:"id"`
	Type       string             `json:"type"`
	Attributes includedAttributes `json:"attributes"`
}

type includedAttributes struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Description string            `json:"description"`
	Duration    float64           `json:"duration"`
	Versions    map[string]string `json:"versions"`
}

const (
	IncludedTypeArtists    = "artists"
	IncludedTypeAudioFiles = "audio_files"
)
