package types

type AppContainer struct {
    AppID     string                 `msgpack:"id"`    // ID Aplikasi (unik)
    Owner     string                 `msgpack:"owner"` // Alamat dompet pengembang
    APIKey    string                 `msgpack:"key"`   // Kunci akses
    Rules     map[string]interface{} `msgpack:"rules"` // JSON Rules ala Firebase
    CreatedAt int64                  `msgpack:"ts"`
}

type UserData struct {
    AppID string `msgpack:"aid"`
    Key   string `msgpack:"k"`
    Value string `msgpack:"v"`
}
