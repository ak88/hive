module github.com/ethereum/hive/simulators/eth2/common

go 1.20

require (
	github.com/ethereum/go-ethereum v1.13.1
	github.com/ethereum/hive v0.0.0-20230401205547-71595beab31d
	github.com/google/uuid v1.3.0
	github.com/herumi/bls-eth-go-binary v1.29.1
	github.com/holiman/uint256 v1.2.3
	github.com/marioevz/eth-clients v0.0.0-20230925172743-e379ee1ecd6e
	github.com/marioevz/mock-builder v1.1.1-0.20230920235938-1f29ea279d7d
	github.com/pkg/errors v0.9.1
	github.com/protolambda/bls12-381-util v0.0.0-20220416220906-d8552aa452c7
	github.com/protolambda/eth2api v0.0.0-20230316214135-5f8afbd6d05d
	github.com/protolambda/go-keystorev4 v0.0.0-20211007151826-f20444f6d564
	github.com/protolambda/zrnt v0.30.0
	github.com/protolambda/ztyp v0.2.2
	github.com/rauljordan/engine-proxy v0.0.0-20230316220057-4c80c36c4c3a
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/wealdtech/go-eth2-util v1.8.1
	golang.org/x/exp v0.0.0-20230810033253-352e893a4cad
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.7.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.9.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v0.0.0-20230906160148-46873a6a7a06 // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.10.0 // indirect
	github.com/crate-crypto/go-kzg-4844 v0.3.0 // indirect
	github.com/deckarep/golang-set/v2 v2.3.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/ethereum/c-kzg-4844 v0.3.1 // indirect
	github.com/ferranbt/fastssz v0.1.3 // indirect
	github.com/getsentry/sentry-go v0.20.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/kilic/bls12-381 v0.1.0 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/supranational/blst v0.3.11 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220614013038-64ee5596c38a // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/wealdtech/go-bytesutil v1.2.1 // indirect
	github.com/wealdtech/go-eth2-types/v2 v2.8.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

replace github.com/protolambda/zrnt => github.com/marioevz/zrnt v0.26.2-0.20230922170744-1bd341bc7f0f

replace github.com/protolambda/eth2api => github.com/marioevz/eth2api v0.0.0-20230922201437-72bd1301e033
