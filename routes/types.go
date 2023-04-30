package routes

type ApiResult struct {
	Error  string      `json:"error"`
	Status string      `json:"status"`
	Result interface{} `json:"result"`
}

type redisResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	ApiResult
}

type redisReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Exp   int    `json:"exp"`
}

type RedisCaseReq struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Exp   int         `json:"exp"`
	Case  string      `json:"case"`
}

type RedisCaseResult struct {
	Key string `json:"key"`
	ApiResult
}

type mongoAggregateResult struct {
	ApiResult
}

type mongoQueryResult struct {
	ApiResult
}

type taskPathsResult struct {
	ApiResult
}

type taskProductDetailReq struct {
	Cmd         string   `json:"cmd"`
	Proxys      []string `json:"proxys"`
	RandomDelay int      `json:"random_delay"`
}

type taskProductDetailResult struct {
	ApiResult
}
