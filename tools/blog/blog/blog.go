package blog

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type post struct {
	config  frontMatter
	content string
}

type frontMatter struct {
	Title     string   `yaml:"title,omitempty"`
	Tags      []string `yaml:"tags,omitempty"`
	Date      string   `yaml:"date,omitempty"`
	Published bool     `yaml:"published"`
}

type replaceMapping struct {
	From string `yaml:"from,omitempty"`
	To   string `yaml:"to,omitempty"`
}

type zennConfig struct {
	PostPath string           `yaml:"post_path,omitempty"`
	Emoji    string           `yaml:"emoji,omitempty"`
	Type     string           `yaml:"type,omitempty"`
	Replaces []replaceMapping `yaml:"replaces,omitempty"`
}

type zennFrontMatter struct {
	Title     string   `yaml:"title,omitempty"`
	Emoji     string   `yaml:"emoji,omitempty"`
	Type      string   `yaml:"type,omitempty"`
	Topics    []string `yaml:"topics,omitempty"`
	Published bool     `yaml:"published"`
}

func readPost(postFilename string) (post, error) {
	b, err := os.ReadFile(postFilename)
	if err != nil {
		return post{}, err
	}

	s := strings.Split(string(b), "---")
	fm := []byte(s[1])
	content := s[2]

	config := frontMatter{}
	if err := yaml.Unmarshal(fm, &config); err != nil {
		return post{}, err
	}

	return post{
		config:  config,
		content: content,
	}, nil
}

func savePost(postFilename string, post post) error {
	out, err := yamlMarshalWithIndent(post.config)
	if err != nil {
		return err
	}
	b := []byte(fmt.Sprintf("---\n%s---%s", out, post.content))
	return os.WriteFile(postFilename, b, 0644)
}

func yamlMarshalWithIndent(v interface{}) (string, error) {
	out := new(bytes.Buffer)
	en := yaml.NewEncoder(out)
	en.SetIndent(2)
	if err := en.Encode(v); err != nil {
		return "", err
	}
	return out.String(), nil
}

func replaceText(content string, replaces []replaceMapping) string {
	for _, r := range replaces {
		content = strings.ReplaceAll(content, r.From, r.To)
	}

	return content
}

func BuildZennArticle(configFilename string) (string, error) {
	b, err := os.ReadFile(configFilename)
	if err != nil {
		return "", err
	}

	config := zennConfig{}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return "", err
	}

	postPath := fmt.Sprintf("content/posts/%s", config.PostPath)
	post, err := readPost(postPath)
	if err != nil {
		return "", err
	}

	post.content = replaceText(post.content, config.Replaces)

	fm := zennFrontMatter{
		Title:     post.config.Title,
		Emoji:     config.Emoji,
		Type:      config.Type,
		Topics:    post.config.Tags,
		Published: post.config.Published,
	}

	out, err := yamlMarshalWithIndent(fm)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("---\n%s---%s", out, post.content), nil
}

func Publish(postFile string) (bool, error) {
	post, err := readPost(postFile)
	if err != nil {
		return false, err
	}

	if post.config.Published {
		return false, nil
	}

	published, err := afterPublicationDate(post.config.Date)
	if err != nil {
		return false, err
	}
	if !published {
		return false, nil
	}
	post.config.Published = true

	return true, savePost(postFile, post)
}

func afterPublicationDate(date string) (bool, error) {
	publishDate, err := time.Parse("2006-01-02T15:04:05+09:00", date)
	if err != nil {
		return false, err
	}
	location := time.FixedZone("Asia/Tokyo", int((9 * time.Hour).Seconds()))
	publishDate = publishDate.In(location).Add(-9 * time.Hour)
	return time.Now().UTC().After(publishDate.UTC()), nil
}

type qiitaPostContentTag struct {
	Name string `json:"name,omitempty"`
}

type qiitaPostContent struct {
	Body    string                `json:"body,omitempty"`
	Private bool                  `json:"private,omitempty"`
	Tags    []qiitaPostContentTag `json:"tags,omitempty"`
	Title   string                `json:"title,omitempty"`
	id      string
	Hash    string
	Edited  bool
}

type qiitaConfig struct {
	PostPath string           `yaml:"post_path,omitempty"`
	Replaces []replaceMapping `yaml:"replaces,omitempty"`
	ID       string           `yaml:"id,omitempty"`
	Hash     string           `yaml:"hash,omitempty"`
}

func UpdateQiitaArticleConf(configFilename, id, hash string) error {
	b, err := os.ReadFile(configFilename)
	if err != nil {
		return err
	}

	config := qiitaConfig{}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return err
	}
	config.ID = id
	config.Hash = hash

	s, err := yamlMarshalWithIndent(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFilename, []byte(s), 0644)
}

func BuildQiitaArticle(configFilename string) (qiitaPostContent, error) {
	b, err := os.ReadFile(configFilename)
	if err != nil {
		return qiitaPostContent{}, err
	}

	config := qiitaConfig{}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return qiitaPostContent{}, err
	}

	postPath := fmt.Sprintf("content/posts/%s", config.PostPath)
	post, err := readPost(postPath)
	if err != nil {
		return qiitaPostContent{}, err
	}

	tags := make([]qiitaPostContentTag, len(post.config.Tags))
	for i, t := range post.config.Tags {
		tags[i].Name = t
	}

	content := qiitaPostContent{
		Title:   post.config.Title,
		Private: !post.config.Published,
		Tags:    tags,
		Body:    replaceText(post.content, config.Replaces),
		id:      config.ID,
	}

	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v %v %v", content.Title, content.Tags, content.Body)))
	content.Hash = fmt.Sprintf("%x", h.Sum(nil))
	content.Edited = content.Hash != config.Hash

	return content, nil
}

func PostArticleToQiita(content qiitaPostContent, authToken string) (string, error) {
	if content.id == "" {
		return postNewArticleToQiita(content, authToken)
	}
	return postEditedArticleToQiita(content, authToken)
}

func postEditedArticleToQiita(content qiitaPostContent, authToken string) (string, error) {
	b, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("https://qiita.com/api/v2/items/%s", content.id),
		bytes.NewBuffer(b),
	)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	req.Header.Add("Content-Type", "application/json")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to post edited article to Qiita. status code is %v", resp.StatusCode)
	}

	return content.id, nil
}

func postNewArticleToQiita(content qiitaPostContent, authToken string) (string, error) {
	b, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest(http.MethodPost, "https://qiita.com/api/v2/items", bytes.NewBuffer(b))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	req.Header.Add("Content-Type", "application/json")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("failed to post new article to Qiita. status code is %v", resp.StatusCode)
	}

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	r := make(map[string]interface{})
	if err := json.Unmarshal(rb, &r); err != nil {
		return "", err
	}

	if v, ok := r["id"]; ok {
		return v.(string), nil
	}

	return "", errors.New("'id' is not contained in responce")
}
