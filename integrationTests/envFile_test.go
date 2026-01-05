package integrationTests

import (
	"os"
	"path"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestEnvFileReadsAllData(t *testing.T) {
	filename := path.Join(t.TempDir(), ".env")

	err := os.WriteFile(
		filename,
		[]byte(`KEY1=value1
KEY2=value2\$3`),
		os.ModePerm,
	)
	assert.Nil(t, err)

	err = godotenv.Load(filename)
	assert.Nil(t, err)

	assert.Equal(t, "value1", os.Getenv("KEY1"))
	assert.Equal(t, "value2$3", os.Getenv("KEY2"))
}
