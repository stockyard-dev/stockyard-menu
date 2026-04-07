package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-menu/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/categories", s.listCategories)
	s.mux.HandleFunc("POST /api/categories", s.createCategories)
	s.mux.HandleFunc("GET /api/categories/export.csv", s.exportCategories)
	s.mux.HandleFunc("GET /api/categories/{id}", s.getCategories)
	s.mux.HandleFunc("PUT /api/categories/{id}", s.updateCategories)
	s.mux.HandleFunc("DELETE /api/categories/{id}", s.delCategories)
	s.mux.HandleFunc("GET /api/items", s.listItems)
	s.mux.HandleFunc("POST /api/items", s.createItems)
	s.mux.HandleFunc("GET /api/items/export.csv", s.exportItems)
	s.mux.HandleFunc("GET /api/items/{id}", s.getItems)
	s.mux.HandleFunc("PUT /api/items/{id}", s.updateItems)
	s.mux.HandleFunc("DELETE /api/items/{id}", s.delItems)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", s.tierHandler)
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	s.mux.HandleFunc("GET /api/extras/{resource}", s.listExtras)
	s.mux.HandleFunc("GET /api/extras/{resource}/{id}", s.getExtras)
	s.mux.HandleFunc("PUT /api/extras/{resource}/{id}", s.putExtras)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"categories": oe(s.db.SearchCategories(q, filters))}); return }
	wj(w, 200, map[string]any{"categories": oe(s.db.ListCategories())})
}

func (s *Server) createCategories(w http.ResponseWriter, r *http.Request) {
	if s.limits.Tier == "none" { we(w, 402, "No license key. Start a 14-day trial at https://stockyard.dev/for/"); return }
	if s.limits.TrialExpired { we(w, 402, "Trial expired. Subscribe at https://stockyard.dev/pricing/"); return }
	var e store.Categories
	json.NewDecoder(r.Body).Decode(&e)
	if e.Name == "" { we(w, 400, "name required"); return }
	s.db.CreateCategories(&e)
	wj(w, 201, s.db.GetCategories(e.ID))
}

func (s *Server) getCategories(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetCategories(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateCategories(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetCategories(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Categories
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.Name == "" { patch.Name = existing.Name }
	s.db.UpdateCategories(&patch)
	wj(w, 200, s.db.GetCategories(patch.ID))
}

func (s *Server) delCategories(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); s.db.DeleteCategories(id); s.db.DeleteExtras("categories", id)
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportCategories(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListCategories()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=categories.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "name", "sort_order", "active", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.Name), fmt.Sprintf("%v", e.SortOrder), fmt.Sprintf("%v", e.Active), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"items": oe(s.db.SearchItems(q, filters))}); return }
	wj(w, 200, map[string]any{"items": oe(s.db.ListItems())})
}

func (s *Server) createItems(w http.ResponseWriter, r *http.Request) {
	var e store.Items
	json.NewDecoder(r.Body).Decode(&e)
	if e.Name == "" { we(w, 400, "name required"); return }
	s.db.CreateItems(&e)
	wj(w, 201, s.db.GetItems(e.ID))
}

func (s *Server) getItems(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetItems(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateItems(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetItems(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Items
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.Name == "" { patch.Name = existing.Name }
	if patch.Category == "" { patch.Category = existing.Category }
	if patch.Description == "" { patch.Description = existing.Description }
	if patch.ImageUrl == "" { patch.ImageUrl = existing.ImageUrl }
	if patch.Dietary == "" { patch.Dietary = existing.Dietary }
	s.db.UpdateItems(&patch)
	wj(w, 200, s.db.GetItems(patch.ID))
}

func (s *Server) delItems(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); s.db.DeleteItems(id); s.db.DeleteExtras("items", id)
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportItems(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListItems()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=items.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "name", "category", "description", "price", "image_url", "dietary", "available", "featured", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.Name), fmt.Sprintf("%v", e.Category), fmt.Sprintf("%v", e.Description), fmt.Sprintf("%v", e.Price), fmt.Sprintf("%v", e.ImageUrl), fmt.Sprintf("%v", e.Dietary), fmt.Sprintf("%v", e.Available), fmt.Sprintf("%v", e.Featured), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["categories_total"] = s.db.CountCategories()
	m["items_total"] = s.db.CountItems()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "menu"}
	m["categories"] = s.db.CountCategories()
	m["items"] = s.db.CountItems()
	wj(w, 200, m)
}

// loadPersonalConfig reads config.json from the data directory.
func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}

// listExtras returns all extras for a resource type as {record_id: {...fields...}}
func (s *Server) listExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	all := s.db.AllExtras(resource)
	out := make(map[string]json.RawMessage, len(all))
	for id, data := range all {
		out[id] = json.RawMessage(data)
	}
	wj(w, 200, out)
}

// getExtras returns the extras blob for a single record.
func (s *Server) getExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	data := s.db.GetExtras(resource, id)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

// putExtras stores the extras blob for a single record.
func (s *Server) putExtras(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	id := r.PathValue("id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		we(w, 400, "read body")
		return
	}
	var probe map[string]any
	if err := json.Unmarshal(body, &probe); err != nil {
		we(w, 400, "invalid json")
		return
	}
	if err := s.db.SetExtras(resource, id, string(body)); err != nil {
		we(w, 500, "save failed")
		return
	}
	wj(w, 200, map[string]string{"ok": "saved"})
}
