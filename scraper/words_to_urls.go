package scraper

import (
	"sync"
)

type WordsToUrls struct {
	mu          sync.RWMutex
	wordsToUrls map[string]map[string]struct{}
}

func NewWordsToUrls() *WordsToUrls {
	return &WordsToUrls{wordsToUrls: make(map[string]map[string]struct{})}
}

func (s *WordsToUrls) Add(word string, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.wordsToUrls[word] == nil {
		s.wordsToUrls[word] = make(map[string]struct{})
	}

	s.wordsToUrls[word][url] = struct{}{}
}

func (s *WordsToUrls) Has(word string, url string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.wordsToUrls[word]

	if !exists {
		return false
	}

	_, exists = s.wordsToUrls[word][url]

	return exists
}

func (s *WordsToUrls) GetItems() map[string]map[string]struct{} {
	return s.wordsToUrls
}
