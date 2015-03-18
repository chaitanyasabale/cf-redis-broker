package backupconfig_test

import (
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/cf-redis-broker/backupconfig"
)

var _ = Describe("Config", func() {
	Describe("Load", func() {
		Context("When the file does not exist", func() {
			It("returns an error", func() {
				_, err := backupconfig.Load("/this/is/an/invalid/path")
				Expect(err.Error()).To(Equal("open /this/is/an/invalid/path: no such file or directory"))
			})
		})

		Context("When the file is successfully loaded", func() {
			var config *backupconfig.Config

			BeforeEach(func() {
				path, err := filepath.Abs(path.Join("assets", "backup.yml"))
				Expect(err).ToNot(HaveOccurred())

				config, err = backupconfig.Load(path)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Has the correct endpoint_url", func() {
				Expect(config.S3Configuration.EndpointUrl).To(Equal("endpointurl.com"))
			})

			It("Has the correct bucket_name", func() {
				Expect(config.S3Configuration.BucketName).To(Equal("some-bucket-name"))
			})

			It("Has the correct access_key_id", func() {
				Expect(config.S3Configuration.AccessKeyId).To(Equal("some-access-key-id"))
			})

			It("Has the correct secret_access_key", func() {
				Expect(config.S3Configuration.SecretAccessKey).To(Equal("secret-access-key"))
			})

			It("Has the correct path", func() {
				Expect(config.S3Configuration.Path).To(Equal("some-s3-path"))
			})

			It("Has the correct region", func() {
				Expect(config.S3Configuration.Region).To(Equal("some-region"))
			})

			It("Has the correct bg_save_timeout", func() {
				Expect(config.BGSaveTimeoutSeconds).To(Equal(10))
			})

			It("Has the correct redis_data_directory", func() {
				Expect(config.RedisDataDirectory).To(Equal("/the/path/to/redis/data"))
			})
		})
	})
})