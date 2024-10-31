package common

import "testing"

func TestValidateNIFPT(t *testing.T) {
	type args struct {
		nif string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Valid NIF 126555397",
			args: args{nif: "126555397"},
			want: true,
		},
		{
			name: "Valid NIF 253557437",
			args: args{nif: "253557437"},
			want: true,
		},
		{
			name: "Valid NIF 303135794",
			args: args{nif: "303135794"},
			want: true,
		},
		{
			name: "Valid NIF 454033206",
			args: args{nif: "454033206"},
			want: true,
		},
		{
			name: "Valid NIF 521649986",
			args: args{nif: "521649986"},
			want: true,
		},
		{
			name: "Valid NIF 649281756",
			args: args{nif: "649281756"},
			want: true,
		},
		{
			name: "Valid NIF 704488272",
			args: args{nif: "704488272"},
			want: true,
		},
		{
			name: "Valid NIF 718453956",
			args: args{nif: "718453956"},
			want: true,
		},
		{
			name: "Valid NIF 727146114",
			args: args{nif: "727146114"},
			want: true,
		},
		{
			name: "Valid NIF 747784540",
			args: args{nif: "747784540"},
			want: true,
		},
		{
			name: "Valid NIF 750633379",
			args: args{nif: "750633379"},
			want: true,
		},
		{
			name: "Valid NIF 771347065",
			args: args{nif: "771347065"},
			want: true,
		},
		{
			name: "Valid NIF 787787825",
			args: args{nif: "787787825"},
			want: true,
		},
		{
			name: "Valid NIF 795768516",
			args: args{nif: "795768516"},
			want: true,
		},
		{
			name: "Valid NIF 800368347",
			args: args{nif: "800368347"},
			want: true,
		},
		{
			name: "Valid NIF 907393314",
			args: args{nif: "907393314"},
			want: true,
		},
		{
			name: "Valid NIF 918620988",
			args: args{nif: "918620988"},
			want: true,
		},
		{
			name: "Valid NIF 985333731",
			args: args{nif: "985333731"},
			want: true,
		},
		{
			name: "Valid NIF 997845775",
			args: args{nif: "997845775"},
			want: true,
		},
		{
			name: "Valid NIF 999999990",
			args: args{nif: "999999990"},
			want: true,
		},
		{
			name: "Invalid NIF 126555398",
			args: args{nif: "126555398"},
			want: false,
		},
		{
			name: "Invalid NIF 253557438",
			args: args{nif: "253557438"},
			want: false,
		},
		{
			name: "Invalid NIF 303135795",
			args: args{nif: "303135795"},
			want: false,
		},
		{
			name: "Invalid NIF 454033207",
			args: args{nif: "454033207"},
			want: false,
		},
		{
			name: "Invalid NIF 521649987",
			args: args{nif: "521649987"},
			want: false,
		},
		{
			name: "Invalid NIF 649281757",
			args: args{nif: "649281757"},
			want: false,
		},
		{
			name: "Invalid NIF 704488273",
			args: args{nif: "704488273"},
			want: false,
		},
		{
			name: "Invalid NIF 718453957",
			args: args{nif: "718453957"},
			want: false,
		},
		{
			name: "Invalid NIF 727146115",
			args: args{nif: "727146115"},
			want: false,
		},
		{
			name: "Invalid NIF 747784541",
			args: args{nif: "747784541"},
			want: false,
		},
		{
			name: "Invalid NIF 750633370",
			args: args{nif: "750633370"},
			want: false,
		},
		{
			name: "Invalid NIF 771347066",
			args: args{nif: "771347066"},
			want: false,
		},
		{
			name: "Invalid NIF 787787826",
			args: args{nif: "787787826"},
			want: false,
		},
		{
			name: "Invalid NIF 795768517",
			args: args{nif: "795768517"},
			want: false,
		},
		{
			name: "Invalid NIF 800368348",
			args: args{nif: "800368348"},
			want: false,
		},
		{
			name: "Invalid NIF 907393315",
			args: args{nif: "907393315"},
			want: false,
		},
		{
			name: "Invalid NIF 918620989",
			args: args{nif: "918620989"},
			want: false,
		},
		{
			name: "Invalid NIF 985333732",
			args: args{nif: "985333732"},
			want: false,
		},
		{
			name: "Invalid NIF 997845776",
			args: args{nif: "997845776"},
			want: false,
		},
		{
			name: "Invalid NIF 999999991",
			args: args{nif: "999999991"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateNIFPT(tt.args.nif); got != tt.want {
				t.Errorf("ValidateNIFPT() = %v, want %v", got, tt.want)
			}
		})
	}
}
