package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Authentication struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
}

type jwtUser struct {
	ID        int    `json:"ID"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"referesh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func (j *Authentication) GeneratToken(user *jwtUser) (TokenPairs, error) {
	//create a token
	token := jwt.New(jwt.SigningMethodHS256)

	//set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["type"] = "JWT"

	//set the expiry
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	//sign the token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	//create a referesh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	//set expiry for refresh token
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()

	//create signed referesh token
	signedRefreshToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	//create tokenpairs and populate with signed token
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	//return tokenpairs
	return tokenPairs, nil
}

func (j *Authentication) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    refreshToken,
		Expires:  time.Now().Add(j.RefreshExpiry),
		MaxAge:   int(j.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func (j *Authentication) GetExpiredRefereshToken() *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func (j *Authentication) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request)(string, *Claims, error){
	w.Header().Add("Vary", "Authorization")

	//get auth header
	authHeader := r.Header.Get("Authorization")
	if authHeader == ""{
		return "", nil, errors.New("no auth header")
	}

	//validate the header
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts)!=2 {
		return "", nil, errors.New("malformed auth header")
	}

	if headerParts[0] != "Bearer"{
		return "", nil, errors.New("malformed auth header")
	}

	//get the token
	token := headerParts[1]
	claims := &Claims{}

	//parse the tokens
	_, err := jwt.ParseWithClaims(token,claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok{
			return nil, fmt.Errorf("unexpected signin method %v",token.Header["Alg"])
		}

		return []byte(j.Secret), nil
	})

	if err!=nil{
		if strings.HasPrefix(err.Error(), "token is expired by"){
			return "", nil, errors.New("expired token")
		}
		return "", nil, errors.New("error here 1")
	}

	//vaidate token issuerer
	if claims.Issuer != j.Issuer{
		return "", nil, errors.New("invalid issuer")
	}

	return token, claims, nil
}