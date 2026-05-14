package service

import "testing"

func TestParseProductImportRowsWithHeader(t *testing.T) {
	text := `品种,牌号,规格,产地,存放地,捆包号,库存,单价
热轧板卷,Q235B,3.0*1500*C,唐钢,唐山仓库A,B-001,52.5,3820
钢管,Q355B,Φ25*6*12000,宝钢,上海仓库,B-002,18,4900`

	items := parseProductImportRows(text)
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	first := items[0].Product
	if first.Category != "热轧板卷" {
		t.Fatalf("category = %q", first.Category)
	}
	if first.Spec.Material != "Q235B" {
		t.Fatalf("material = %q", first.Spec.Material)
	}
	if first.Spec.HeightSpecValue != 3 {
		t.Fatalf("height spec = %v", first.Spec.HeightSpecValue)
	}
	if first.Spec.WidthSpecValue != 1500 {
		t.Fatalf("width spec = %v", first.Spec.WidthSpecValue)
	}
	if first.Spec.LengthType != "coil" {
		t.Fatalf("length type = %q", first.Spec.LengthType)
	}
	if first.Quantity != 52.5 || first.UnitPrice != 3820 {
		t.Fatalf("quantity/unit price = %v/%v", first.Quantity, first.UnitPrice)
	}

	second := items[1].Product
	if second.Spec.HeightSpecType != "diameter" {
		t.Fatalf("pipe height spec type = %q", second.Spec.HeightSpecType)
	}
	if second.Spec.WidthSpecType != "wall_thickness" {
		t.Fatalf("pipe width spec type = %q", second.Spec.WidthSpecType)
	}
	if second.Spec.LengthValue != 12000 {
		t.Fatalf("pipe length value = %v", second.Spec.LengthValue)
	}
}

func TestParseProductImportRowsWithoutHeader(t *testing.T) {
	text := "唐钢 Q235B 热轧板卷 3.0*1500*C 库存50 单价3800"

	items := parseProductImportRows(text)
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	item := items[0].Product
	if item.Spec.Material != "Q235B" {
		t.Fatalf("material = %q", item.Spec.Material)
	}
	if item.Category != "热轧板卷" {
		t.Fatalf("category = %q", item.Category)
	}
	if item.SteelMill != "唐钢" {
		t.Fatalf("steel mill = %q", item.SteelMill)
	}
	if item.Quantity != 50 || item.UnitPrice != 3800 {
		t.Fatalf("quantity/unit price = %v/%v", item.Quantity, item.UnitPrice)
	}
}
