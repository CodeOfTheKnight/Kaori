package main

type endpoints string

const (
	endpointRoot endpoints = "/"
	endpointGui  endpoints = "/KaoriGui/"

	//Endpoint per servizio
	endApiAnime   endpoints = "/api/anime"   //[GET] Parametri: IdAnilist
	endApiManga   endpoints = "/api/manga"   //[GET] Parametri: IdAnilist
	endApiChapter endpoints = "/api/chapter" //[GET] Parametri: IdAnilist, numEpisodio, sito
	endApiEpisode endpoints = "/api/episode" //[GET] Parametri: IdAnilist, numCapitolo, sito
	endApiMuisc   endpoints = "/api/music"   //[GET] Parametri: IdAnilist, [tipo](OP, ED, OST)

	//Endpoint utenti
	endApiSignUp endpoints = "/api/user/signup" //[POST] Parametri: Username, mail, password.
	endApiLogin  endpoints = "/api/user/login"  //[POST] Parametri: mail, password  | Utilizza basic auth + JWT

	//Endpoint addData
	endApiAddData      endpoints = "/api/addData/"
	endApiAddDataAnime endpoints = "/anime" //[POST] Struttura dati da definire.
	endApiAddDataManga endpoints = "/manga" //[POST] Struttura dati da definire.
	endApiAddDataMusic endpoints = "/music" //[POST] Struttura dati da definire.

	//Endpoint lista anime/manga
	endpointUserList endpoints = "/api/list/userList" //[Get] Parametri: TokenAnilist, TokenServer, TipoLista
	endpointUserInfo endpoints = "/api/list/userinfo" //[Get] Parametri: TokenAnilist, TokenServer

	//Endpoint for tests
	endpointTest endpoints = "/api/test/"
	testFiles endpoints = "/files/"
)

func (e endpoints) String() string {
	return string(e)
}
