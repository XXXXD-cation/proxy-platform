package webshare

// Response 是 Webshare API /proxy/list/ 端点返回的JSON对象的顶层结构。
type Response struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []Proxy `json:"results"`
}

// Proxy 代表 Webshare API 返回的代理列表中的单个代理对象。
type Proxy struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	ProxyAddress     string `json:"proxy_address"`
	Port             uint16 `json:"port"`
	Valid            bool   `json:"valid"`
	LastVerification string `json:"last_verification"`
	CountryCode      string `json:"country_code"`
	CityName         string `json:"city_name"`
	CreatedAt        string `json:"created_at"`
}
