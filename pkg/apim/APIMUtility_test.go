package apim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOdataQueryInitExisting(t *testing.T) {
	t.Run("MakeOdataQuery- Success Test", func(t *testing.T) {
		odataFilterInputs := OdataParameters{Filter: "isCopy eq false", Search: "",
			Top: 4, Skip: 1, Orderby: "name",
			Select: "", Expand: ""}
		odataFilters, err := OdataUtils.MakeOdataQuery(&odataFilterInputs)
		assert.NoError(t, err)
		assert.Equal(t, "?filter=isCopy+eq+false&$orderby=name&$skip=1&$top=4", odataFilters)
	})

	t.Run("MakeOdataQuery- empty odata filters Test", func(t *testing.T) {
		odataFilterInputs := OdataParameters{Filter: "", Search: "",
			Top: 0, Skip: 0, Orderby: "",
			Select: "", Expand: ""}
		odataFilters, err := OdataUtils.MakeOdataQuery(&odataFilterInputs)
		assert.NoError(t, err)
		assert.Equal(t, "", odataFilters)
	})
}
