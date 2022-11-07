package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	r2 := R2{}
	r2.Setup()

	args := os.Args[1:]
	for _, img := range args {
		if strings.HasPrefix(img, "http") {
			file, err := os.CreateTemp(os.TempDir(), "r2_uploader_*_"+filepath.Ext(img))
			if err != nil {
				log.Fatalf("Failed to create temp file: %v", err)
			}

			resp, err := http.Get(img)
			if err != nil {
				log.Fatalf("Failed to download image: %v", err)
			}

			_, err = io.Copy(file, resp.Body)
			if err != nil {
				log.Fatalf("Failed to copy image: %v", err)
			}

			_ = resp.Body.Close()
			_ = file.Close()

			img = file.Name()
		}

		fmt.Println(r2.Upload(&r2.config, img))
	}
}

type Config struct {
	AccountID  string `json:"account_id"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	BucketName string `json:"bucket_name"`
	PublicURL  string `json:"public_url"`
}

func (c *Config) getConfigDir() string {
	var baseDir string
	var configDir string

	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("APPDATA")
		configDir = "r2_uploader"
	} else if runtime.GOOS == "linux" {
		baseDir = os.Getenv("HOME")
		configDir = ".config/r2_uploader"
	} else {
		log.Fatalf("Unsupported OS: %s", runtime.GOOS)
	}

	if baseDir == "" {
		baseDir = "."
	}

	dir := path.Join(baseDir, configDir)
	_ = os.MkdirAll(dir, 0755)

	return dir
}

func (c *Config) Load() {
	data, err := os.ReadFile(filepath.Join(c.getConfigDir(), "config.json"))
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	if c.AccountID == "" || c.AccessKey == "" || c.SecretKey == "" || c.BucketName == "" {
		log.Fatalf("Config file is missing required fields")
	}
}

type R2 struct {
	config Config
	client *s3.Client
}

func (r *R2) randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)
}

func (r *R2) Setup() {
	r.config = Config{}
	r.config.Load()

	endpoint := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r.config.AccountID),
			SigningRegion: "auto",
		}, nil
	})
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("auto"),
		config.WithEndpointResolverWithOptions(endpoint),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r.config.AccessKey, r.config.SecretKey, "")),
	)
	if err != nil {
		log.Fatalf("Failed to load r2 config: %v", err)
		return
	}

	r.client = s3.NewFromConfig(cfg)
}

func (r *R2) Upload(c *Config, img string) string {
	stat, err := os.Stat(img)
	if err != nil {
		log.Fatalf("Failed to stat image: %v", err)
	}

	file, err := os.Open(img)
	if err != nil {
		log.Fatalf("Failed to open image: %v", err)
	}

	key := fmt.Sprintf("images/%s/%s%s", time.Now().Format("2006/01/02"), r.randomString(16), filepath.Ext(img))

	_, err = r.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(c.BucketName),
		Key:           aws.String(key),
		Body:          file,
		ContentLength: stat.Size(),
	})
	_ = file.Close()
	if err != nil {
		log.Fatalf("Failed to upload image: %v", err)
	}

	return c.PublicURL + "/" + key
}
