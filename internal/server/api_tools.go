package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/tanq16/anbu/internal/tools"
)

const maxUUIDs = 100

func (s *Server) routeTools() {
	s.mux.HandleFunc("POST /api/tools/hash", s.handleHash)
	s.mux.HandleFunc("POST /api/tools/yaml", s.handleYAML)
	s.mux.HandleFunc("POST /api/tools/time", s.handleTime)
	s.mux.HandleFunc("GET /api/tools/uuid", s.handleUUID)
	s.mux.HandleFunc("GET /api/tools/passphrase", s.handlePassphrase)
	s.mux.HandleFunc("GET /api/tools/random", s.handleRandom)
}

func (s *Server) handleHash(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tools.HashAll([]byte(in.Text)))
}

func (s *Server) handleYAML(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Input string `json:"input"`
		To    string `json:"to"`
	}
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	var out string
	var err error
	switch in.To {
	case "json":
		out, err = tools.YAMLToJSON(in.Input)
	case "yaml":
		out, err = tools.JSONToYAML(in.Input)
	default:
		s.fail(w, r, badRequest(`to must be "json" or "yaml"`))
		return
	}
	if err != nil {
		s.fail(w, r, badRequest("%v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"output": out})
}

func (s *Server) handleTime(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Op     string  `json:"op"`
		Input  string  `json:"input"`
		Epochs []int64 `json:"epochs"`
	}
	if err := readJSON(w, r, &in); err != nil {
		s.fail(w, r, err)
		return
	}
	switch in.Op {
	case "now":
		writeJSON(w, http.StatusOK, tools.FormatTime(time.Now()))
	case "parse":
		t, err := tools.ParseTime(in.Input)
		if err != nil {
			s.fail(w, r, badRequest("%v", err))
			return
		}
		writeJSON(w, http.StatusOK, tools.FormatTime(t))
	case "until":
		target, now, err := tools.TimeUntil(in.Input)
		if err != nil {
			s.fail(w, r, badRequest("%v", err))
			return
		}
		writeJSON(w, http.StatusOK, tools.Until(target, now))
	case "diff":
		if len(in.Epochs) < 1 || len(in.Epochs) > 2 {
			s.fail(w, r, badRequest("epochs takes one or two values"))
			return
		}
		writeJSON(w, http.StatusOK, tools.EpochDiff(tools.TimeEpochDiff(in.Epochs)))
	default:
		s.fail(w, r, badRequest("op must be now, parse, until, or diff"))
	}
}

func queryInt(r *http.Request, key string, def int) (int, error) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, badRequest("%s must be a number", key)
	}
	return n, nil
}

func queryBool(r *http.Request, key string, def bool) (bool, error) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, badRequest("%s must be true or false", key)
	}
	return b, nil
}

func (s *Server) handleUUID(w http.ResponseWriter, r *http.Request) {
	version, err := queryInt(r, "version", s.settings.Get().UUIDVersion)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	count, err := queryInt(r, "count", 1)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	short, err := queryBool(r, "short", false)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if version != 4 && version != 7 {
		s.fail(w, r, badRequest("version must be 4 or 7"))
		return
	}
	if count < 1 || count > maxUUIDs {
		s.fail(w, r, badRequest("count must be between 1 and %d", maxUUIDs))
		return
	}
	values := make([]string, 0, count)
	for range count {
		gen := tools.GenerateUUIDString
		if short {
			gen = tools.GenerateShortUUIDString
		}
		v, err := gen(version == 4)
		if err != nil {
			s.fail(w, r, err)
			return
		}
		values = append(values, v)
	}
	writeJSON(w, http.StatusOK, map[string][]string{"values": values})
}

func (s *Server) handlePassphrase(w http.ResponseWriter, r *http.Request) {
	cur := s.settings.Get()
	words, err := queryInt(r, "words", cur.PassphraseWords)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	simple, err := queryBool(r, "simple", cur.PassphraseSimple)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if words < 1 || words > 50 {
		s.fail(w, r, badRequest("words must be between 1 and 50"))
		return
	}
	value, err := tools.GeneratePassPhrase(words, simple)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"value": value})
}

func (s *Server) handleRandom(w http.ResponseWriter, r *http.Request) {
	cur := s.settings.Get()
	length, err := queryInt(r, "length", cur.RandomLength)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if length < 1 || length > 1024 {
		s.fail(w, r, badRequest("length must be between 1 and 1024"))
		return
	}
	name := r.URL.Query().Get("charset")
	if name == "" {
		name = cur.RandomCharset
	}
	charset, ok := tools.Charsets[name]
	if !ok {
		s.fail(w, r, badRequest("charset must be one of alphanumeric, alpha, digits, hex, all"))
		return
	}
	value, err := tools.GenerateRandomStringCharset(length, charset)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"value": value})
}
