package handlers

import (
	"crypto/rand"
	_ "embed"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/datasektionen/sso/database"
	"github.com/datasektionen/sso/pkg/email"
	"github.com/datasektionen/sso/pkg/httputil"
	"github.com/datasektionen/sso/service"
	"github.com/datasektionen/sso/templates"
)

var (
	//go:embed emaillogin_words.txt
	wordsRaw  string
	idxToWord []string
	wordToIdx = map[string]int{}
)

func init() {
	for _, word := range strings.Split(wordsRaw, "\n") {
		if len(word) == 0 {
			break
		}
		i := len(idxToWord)
		idxToWord = append(idxToWord, word)
		wordToIdx[word] = i
	}
	if len(idxToWord) != 2048 {
		panic("Expected 2048 words in wordlist")
	}
}

func BytesTo11BitInts(bytes []byte) func(func(uint16) bool) {
	return func(yield func(uint16) bool) {
		buf, bits := uint32(0), 0
		for _, b := range bytes {
			buf = buf | (uint32(b) << bits)
			bits += 8
			for bits >= 11 {
				bits -= 11
				idx := buf & 0b11111111111
				buf = buf >> 11
				if !yield(uint16(idx)) {
					return
				}
			}
		}
	}
}

// randomCode generates a code of 11 english words from a wordlist of 2048 words, thus it has 121 bits of entropy
func randomCode() string {
	var randBytes [16]byte
	rand.Read(randBytes[:])

	var code string
	for idx := range BytesTo11BitInts(randBytes[:]) {
		if code != "" {
			code += " "
		}
		code += idxToWord[idx]
	}

	return code
}

func parseCode(text string) (string, error) {
	var res string
	count := 0
	for _, word := range strings.Split(text, " ") {
		word := strings.TrimSpace(word)
		if len(word) == 0 {
			continue
		}
		_, ok := wordToIdx[word]
		if !ok {
			return "", httputil.BadRequest("'" + word + "' is not in the wordlist, so this cannot be a valid code")
		}
		if len(res) != 0 {
			res += " "
		}
		res += word
		count += 1
	}
	if count != 11 {
		return "", httputil.BadRequest("The code always consists of 11 words. You provided " + strconv.Itoa(count))
	}
	return res, nil
}

func beginLoginEmail(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	kthid := r.FormValue("kthid")
	user, err := s.GetUser(r.Context(), kthid)
	if err != nil {
		return err
	}
	if user == nil {
		return httputil.BadRequest("No such user")
	}

	code := randomCode()

	if err := s.DB.BeginEmailLogin(
		r.Context(),
		database.BeginEmailLoginParams{Kthid: kthid, Code: code},
	); err != nil {
		return err
	}

	if err := email.Send(
		r.Context(),
		user.Email,
		"SSO - Login Code",
		`Your temporary login code is `+code,
	); err != nil {
		return err
	}

	return templates.CodeForm(kthid)
}

func finishLoginEmail(s *service.Service, w http.ResponseWriter, r *http.Request) httputil.ToResponse {
	kthid := r.FormValue("kthid")
	code := r.FormValue("code")

	code, err := parseCode(code)
	if err != nil {
		return err
	}

	res, err := s.DB.FinishEmailLogin(r.Context(), database.FinishEmailLoginParams{
		Kthid: kthid,
		Code:  code,
	})
	if err != nil {
		return err
	}

	if res.Ok {
		return s.LoginUser(r.Context(), kthid, true)
	} else {
		slog.Info("Failed email login", "kthid", kthid, "code", code, "reason", res.Reason)
		msg := "Code is invalid for an unkown reason. (please tell d-sys)"
		switch res.Reason {
		case "expired":
			msg = "Login code has expired. Please restart."
		case "exhausted":
			msg = "Too many invalid attempts. Please restart."
		case "wrong":
			msg = "Invalid code. Please copy-paste or spell better."
		case "no code":
			msg = "User does not have a login code. Please restart."
		}
		return httputil.BadRequest(msg)
	}
}
