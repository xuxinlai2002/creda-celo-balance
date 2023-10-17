package tokens

type TokenInfo struct {
	Name     string
	CoinID   uint64
	Decimals uint8
}

var ERC20Tokens = map[string]TokenInfo{
	"0x617f3112bf5397D0467D315cC709EF968D9ba546": TokenInfo{Name: "USDT", CoinID: 825, Decimals: 6},
	"0xef4229c8c3250C675F21BCefa42f58EfbfF6002a": TokenInfo{Name: "USDC", CoinID: 3408, Decimals: 6},
	"0x37f750B7cC259A2f741AF45294f6a16572CF5cAd": TokenInfo{Name: "USDC(WormHole)", CoinID: 20650, Decimals: 6},
	"0xD629eb00dEced2a080B7EC630eF6aC117e614f1b": TokenInfo{Name: "WBTC", CoinID: 3717, Decimals: 18},
	"0x471EcE3750Da237f93B8E339c536989b8978a438": TokenInfo{Name: "CELO", CoinID: 5567, Decimals: 18},
	"0x29dFce9c22003A4999930382Fd00f9Fd6133Acd1": TokenInfo{Name: "SUSHI", CoinID: 6758, Decimals: 18},
	"0xB9C8F0d3254007eE4b98970b94544e473Cd610EC": TokenInfo{Name: "MIMATIC", CoinID: 10238, Decimals: 18},
	"0xD8763CBa276a3738E6DE85b4b3bF5FDed6D6cA73": TokenInfo{Name: "cEUR", CoinID: 9467, Decimals: 18},
	"0x9995cc8F20Db5896943Afc8eE0ba463259c931ed": TokenInfo{Name: "ETHIX", CoinID: 8442, Decimals: 18},
	"0x765DE816845861e75A25fCA122bb6898B8B1282a": TokenInfo{Name: "cUSD", CoinID: 7236, Decimals: 18},
	"0x1d18d0386F51ab03E7E84E71BdA1681EbA865F1f": TokenInfo{Name: "JMPT", CoinID: 17334, Decimals: 18},
	"0x27cd006548dF7C8c8e9fdc4A67fa05C2E3CA5CF9": TokenInfo{Name: "PLASTIK", CoinID: 15575, Decimals: 9},
	"0xEe9801669C6138E84bD50dEB500827b776777d28": TokenInfo{Name: "O3", CoinID: 9588, Decimals: 18},
	"0x6e512BFC33be36F2666754E996ff103AD1680Cc9": TokenInfo{Name: "ABR", CoinID: 12212, Decimals: 18},
	"0x00Be915B9dCf56a3CBE739D9B9c202ca692409EC": TokenInfo{Name: "UBE", CoinID: 10808, Decimals: 18},
	"0x17700282592D6917F6A73D0bF8AcCf4D578c131e": TokenInfo{Name: "MOO", CoinID: 13021, Decimals: 18},
	"0xe8537a3d056DA446677B9E9d6c5dB704EaAb4787": TokenInfo{Name: "CREAL", CoinID: 16385, Decimals: 18},
}

var TokenDecimals = map[string]int{
	"825":   6,
	"3408":  6,
	"20650": 6,
	"3717":  18,
	"5567":  18,
	"6758":  18,
	"10238": 18,
	"9467":  18,
	"8442":  18,
	"7236":  18,
	"17334": 18,
	"15575": 9,
	"9588":  18,
	"12212": 18,
	"10808": 18,
	"13021": 18,
	"16385": 18,
}
