package harvester_test

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/mgjules/harvit/converter"
	"github.com/mgjules/harvit/harvester"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Website", func() {
	b, err := ioutil.ReadFile("testdata/website.html")
	Expect(err).To(BeNil())

	ts := httptest.NewServer(writeHTML(string(b)))
	AfterEach(func() {
		ts.Close()
	})

	p := plan.Plan{
		Source: ts.URL,
		Type:   harvester.TypeWebsite,
		Fields: []plan.Field{
			{
				Name:     "raw",
				Type:     converter.TypeRaw,
				Selector: "#app > p.raw",
			},
			{
				Name:     "text",
				Type:     converter.TypeText,
				Selector: "#app > p.text",
			},
			{
				Name:     "textList",
				Type:     converter.TypeText,
				Selector: "#app > ul.text-list > li",
			},
			{
				Name:     "number",
				Type:     converter.TypeNumber,
				Selector: "#app > p.number",
			},
			{
				Name:     "numberWithText",
				Type:     converter.TypeNumber,
				Selector: "#app > p.number-with-text",
			},
			{
				Name:     "decimal",
				Type:     converter.TypeDecimal,
				Selector: "#app > p.decimal",
			},
			{
				Name:     "decimalWithText",
				Type:     converter.TypeDecimal,
				Selector: "#app > p.decimal-with-text",
			},
			{
				Name:     "datetime",
				Type:     converter.TypeDateTime,
				Selector: "#app > p.datetime",
			},
			{
				Name:     "datetimeWithText",
				Type:     converter.TypeDateTime,
				Selector: "#app > p.datetime-with-text",
			},
		},
	}

	expected := map[string]any{
		"raw":  `<p class="raw">Get html!</p>`,
		"text": "Some t3xt!",
		"textList": []string{
			"1Sw0C0tlYNfC2ookd5lr",
			"ifpTMDlSfhMSCD",
			"kRaQ5Lqtrbrk1oEq",
			"Q9g17hjV",
			"hUPwfr1GKzaHkMmENn",
		},
		"number":           "1337",
		"numberWithText":   "This is some leet number: 1337",
		"decimal":          "13.37",
		"decimalWithText":  "This is some leet decimal: 13.37",
		"datetime":         "08/06/2022 19:53:44",
		"datetimeWithText": "This is some random datetime: 08/06/2022 19:53:44",
	}

	_, err = logger.New(false)
	Expect(err).To(BeNil())

	h, err := harvester.New(p.Type)
	Expect(err).To(BeNil())

	It("should successfully harvest the data", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		data, err := h.Harvest(ctx, &p)
		Expect(err).To(BeNil())

		Expect(data).To(BeEquivalentTo(expected))
	})
})

func writeHTML(content string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, strings.TrimSpace(content))
	})
}
