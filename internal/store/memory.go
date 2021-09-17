package store

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/gofrs/uuid"
)

var _ Store = (*InMemory)(nil)
var _ AuthStore = (*InMemory)(nil)

type InMemory struct {
	store     map[string]*url.URL
	userStore map[string]map[string]*url.URL
}

// NewInMemory create new InMemory instance
func NewInMemory() *InMemory {
	return &InMemory{
		store:     make(map[string]*url.URL),
		userStore: make(map[string]map[string]*url.URL),
	}
}

func (m *InMemory) Save(_ context.Context, u *url.URL) (id string, err error) {
	id = fmt.Sprintf("%x", len(m.store))
	m.store[id] = u
	return id, nil
}

func (m *InMemory) SaveBatch(_ context.Context, urls []*url.URL) (ids []string, err error) {
	ids = make([]string, 0, len(urls))
	for _, u := range urls {
		id := fmt.Sprintf("%x", len(m.store))
		m.store[id] = u
		ids = append(ids, id)
	}
	if len(ids) != len(urls) {
		return nil, errors.New("not all URLs have been saved")
	}
	return
}

func (m *InMemory) Load(_ context.Context, id string) (u *url.URL, err error) {
	if u, ok := m.store[id]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (m *InMemory) SaveUser(ctx context.Context, uid uuid.UUID, u *url.URL) (id string, err error) {
	id, err = m.Save(ctx, u)
	if err != nil {
		return "", fmt.Errorf("cannot save URL to shared store: %w", err)
	}
	if _, ok := m.userStore[uid.String()]; !ok {
		m.userStore[uid.String()] = make(map[string]*url.URL)
	}
	m.userStore[uid.String()][id] = u
	return id, nil
}

func (m *InMemory) SaveUserBatch(ctx context.Context, uid uuid.UUID, urls []*url.URL) (ids []string, err error) {
	ids, err = m.SaveBatch(ctx, urls)
	if err != nil {
		return nil, fmt.Errorf("cannot save URLs to shared store: %w", err)
	}
	if _, ok := m.userStore[uid.String()]; !ok {
		m.userStore[uid.String()] = make(map[string]*url.URL)
	}
	for i, id := range ids {
		m.userStore[uid.String()][id] = urls[i]
	}
	return ids, nil
}

func (m *InMemory) LoadUser(ctx context.Context, uid uuid.UUID, id string) (u *url.URL, err error) {
	urls, err := m.LoadUsers(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("cannot load user urls: %w", err)
	}
	if u, ok := urls[id]; ok {
		return u, nil
	}
	return nil, ErrNotFound
}

func (m *InMemory) LoadUsers(_ context.Context, uid uuid.UUID) (urls map[string]*url.URL, err error) {
	if urls, ok := m.userStore[uid.String()]; ok {
		return urls, nil
	}
	return nil, ErrNotFound
}

func (m *InMemory) Close() error {
	return nil
}

func (m *InMemory) Ping(_ context.Context) error {
	return nil
}