package common

import (
	"fmt"
	"io"
	"slices"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var (
	_ io.Writer = &LogCapture{}
)

type LogCapture struct {
	Logs []string
	mu   sync.Mutex
}

func (w *LogCapture) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Logs = append(w.Logs, string(p))
	return len(p), nil
}

func TestValidateStructTags(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		path           string
		expectedErr    error
		expectedLogs   []string // Expected log messages
		unexpectedLogs []string // Log messages that should not appear
	}{
		// {
		// 	name: "Valid struct with no deprecated fields does not warn",
		// 	input: struct {
		// 		Field1 string `yaml:"field1" json:"field1"`
		// 		Field2 int    `yaml:"field2" json:"field2"`
		// 	}{
		// 		Field1: "test",
		// 		Field2: 1,
		// 	},
		// 	path:           "TestStruct",
		// 	expectedErr:    nil,
		// 	expectedLogs:   []string{},
		// 	unexpectedLogs: []string{"deprecated", "removed", "replaced"},
		// },
		// {
		// 	name: "Struct with zeroed deprecated fields does not warn",
		// 	input: struct {
		// 		Field1 string `yaml:"field1" json:"field1" deprecated:"removed"`
		// 		Field2 int    `yaml:"field2" json:"field2" deprecated:"replaced,newField"`
		// 	}{},
		// 	path:           "TestStruct",
		// 	expectedErr:    nil,
		// 	expectedLogs:   []string{},
		// 	unexpectedLogs: []string{"deprecated", "removed", "replaced"},
		// },
		{
			name: "Struct with deprecated fields warns",
			input: struct {
				Field1 string `yaml:"field1" json:"field1" deprecated:"removed"`
				Field2 string `yaml:"field2" json:"field2" deprecated:"replaced,newField"`
			}{
				Field1: "test",
				Field2: "test",
			},
			path:        "TestStruct",
			expectedErr: nil,
			expectedLogs: []string{
				"This field will be removed in a future version",
				"This field is deprecated and will be removed in a future version. Please use 'newField' instead",
			},
			unexpectedLogs: []string{},
		},
		// {
		// 	name: "Struct with nested deprecated fields",
		// 	input: struct {
		// 		Nested struct {
		// 			Field1 string `yaml:"field1" json:"field1" deprecated:"removed"`
		// 		} `yaml:"nested" json:"nested"`
		// 	}{},
		// 	path:     "TestStruct",
		// 	expected: nil,
		// 	expectedLogs: []string{
		// 		"This field will be removed in a future version",
		// 	},
		// 	unexpectedLogs: []string{},
		// },
		// {
		// 	name: "Struct with slice of deprecated fields",
		// 	input: struct {
		// 		Fields []struct {
		// 			Field1 string `yaml:"field1" json:"field1" deprecated:"removed"`
		// 		} `yaml:"fields" json:"fields"`
		// 	}{},
		// 	path:     "TestStruct",
		// 	expected: nil,
		// 	expectedLogs: []string{
		// 		"This field will be removed in a future version",
		// 	},
		// 	unexpectedLogs: []string{},
		// },
		// {
		// 	name: "Struct with map of deprecated fields",
		// 	input: struct {
		// 		Fields map[string]struct {
		// 			Field1 string `yaml:"field1" json:"field1" deprecated:"removed"`
		// 		} `yaml:"fields" json:"fields"`
		// 	}{},
		// 	path:     "TestStruct",
		// 	expected: nil,
		// 	expectedLogs: []string{
		// 		"This field will be removed in a future version",
		// 	},
		// 	unexpectedLogs: []string{},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a custom writer to capture the output
			capture := &LogCapture{}
			logger := zerolog.New(capture).With().Timestamp().Logger().Level(zerolog.WarnLevel)

			// Test that logging works at all
			logger.Info().Msg("test info message")
			logger.Warn().Msg("test warning message")
			fmt.Println("Buffer contents after test messages:", capture.Logs)

			err := validateStructTags(tt.input, tt.path, &logger)
			assert.Equal(t, tt.expectedErr, err)

			unexpectedLogs := []string{}
			seenLogs := map[string]bool{}
			for _, log := range capture.Logs {
				fmt.Println(log)

				if !slices.Contains(tt.expectedLogs, log) && !seenLogs[log] {
					unexpectedLogs = append(unexpectedLogs, log)
					seenLogs[log] = true
				}
			}

			assert.Equal(t, len(tt.expectedLogs), len(seenLogs))
			assert.Empty(t, unexpectedLogs)
		})
	}
}
