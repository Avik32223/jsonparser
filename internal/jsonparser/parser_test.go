package jsonparser

import (
	"os"
	"testing"
)

func Test_parser_Parse(t *testing.T) {
	type testCase struct {
		name     string
		filename string
		wantErr  bool
	}

	tests := []testCase{
		{"step1-invalid-1", "./testcases/step1/invalid.json", true},
		{"step1-valid-1", "./testcases/step1/valid.json", false},
		{"step2-invalid-1", "./testcases/step2/invalid.json", true},
		{"step2-invalid-2", "./testcases/step2/invalid2.json", true},
		{"step2-valid-1", "./testcases/step2/valid.json", false},
		{"step2-valid-2", "./testcases/step2/valid2.json", false},
		{"step3-valid-1", "./testcases/step3/valid.json", false},
		{"step3-invalid-1", "./testcases/step3/invalid.json", true},
		{"step4-valid-1", "./testcases/step4/valid.json", false},
		{"step4-valid-2", "./testcases/step4/valid2.json", false},
		{"step4-invalid-1", "./testcases/step4/invalid.json", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Errorf(err.Error())
			}
			p := &parser{
				source: string(f),
			}
			_, err = p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
