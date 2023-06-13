package types_test

import (
	time "time"

	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibcgptypes "github.com/cosmos/ibc-go/v6/modules/light-clients/10-grandpa/types"
	"github.com/dablelv/go-huge-util/conv"
	"github.com/octopus-network/beefy-go/beefy"
)

var gpClientState = ibcgptypes.ClientState{
	ChainType:         0,
	ChainId:           chainID,
	ParachainId:       0,
	LatestBeefyHeight: clienttypes.NewHeight(clienttypes.ParseChainID(chainID), 15228),
	LatestMmrRoot:     conv.SplitStrToSlice[byte]("131 79 104 195 33 161 208 242 156 164 3 120 80 122 102 198 67 105 240 96 40 47 16 197 136 94 190 101 145 9 176 52", " "),
	LatestChainHeight: clienttypes.NewHeight(clienttypes.ParseChainID(chainID), 15228),
	FrozenHeight:      clienttypes.NewHeight(clienttypes.ParseChainID(chainID), 0),
	LatestAuthoritySet: ibcgptypes.BeefyAuthoritySet{
		Id:   1522,
		Len:  5,
		Root: conv.SplitStrToSlice[byte]("48 72 3 250 90 145 217 133 44 170 254 4 180 184 103 164 237 39 160 122 91 238 61 21 7 180 177 135 166 135 119 162", " "),
	},
}

var payloads = []ibcgptypes.PayloadItem{
	{
		Id:   conv.SplitStrToSlice[byte]("109 104", " "),
		Data: conv.SplitStrToSlice[byte]("31 52 208 2 3 182 252 155 182 210 209 187 127 178 123 44 217 192 102 62 47 189 24 87 59 135 216 57 171 69 148 102", " "),
	},
}
var commitment = ibcgptypes.Commitment{
	Payloads:       payloads,
	BlockNumber:    15230,
	ValidatorSetId: 1523,
}
var sinatures = []ibcgptypes.Signature{
	{
		Index:     0,
		Signature: conv.SplitStrToSlice[byte]("242 44 55 231 203 28 113 208 253 193 158 42 182 13 198 126 73 157 182 199 226 71 42 194 1 72 234 39 54 88 12 209 69 163 82 239 83 5 198 73 180 95 50 202 7 48 212 44 218 109 102 173 171 226 84 222 209 219 249 135 104 241 54 215 1", " "),
	},
	{
		Index:     1,
		Signature: conv.SplitStrToSlice[byte]("191 117 252 169 82 231 224 16 131 2 201 160 111 180 214 10 20 90 172 201 238 160 13 232 117 214 46 219 49 201 66 43 21 133 167 194 146 9 44 122 85 247 235 221 35 215 247 165 55 124 237 145 163 220 116 181 39 150 90 19 8 172 129 42 0", " "),
	},
	{
		Index:     2,
		Signature: conv.SplitStrToSlice[byte]("19 244 0 196 122 249 183 17 159 40 109 17 179 151 220 208 119 213 143 145 105 106 69 66 118 89 235 66 52 51 60 93 103 24 58 204 129 70 63 107 235 170 12 187 71 230 81 152 2 236 184 128 43 207 69 222 91 251 132 186 195 201 168 194 1", " "),
	},
	{
		Index:     4,
		Signature: conv.SplitStrToSlice[byte]("222 108 146 166 109 244 200 92 67 70 13 220 177 30 53 162 73 73 172 88 114 218 79 9 64 20 165 197 234 122 112 161 109 123 46 205 237 123 116 25 28 67 151 102 209 205 31 89 248 123 159 78 166 222 69 57 192 35 43 9 251 57 189 140 1", " "),
	},
}

var signedCommitment = ibcgptypes.SignedCommitment{
	Commitment: commitment,
	Signatures: sinatures,
}

var signatureProofs = [][]byte{
	conv.SplitStrToSlice[byte]("80 189 211 172 79 84 160 71 2 160 85 195 51 3 2 91 32 56 68 108 115 52 237 59 51 65 243 16 240 82 17 111", " "),
}
var leaves = []ibcgptypes.MMRLeaf{
	{
		Version: 0,
		ParentNumberAndHash: ibcgptypes.ParentNumberAndHash{
			ParentNumber: 15228,
			ParentHash:   conv.SplitStrToSlice[byte]("242 17 28 1 162 114 136 220 85 166 164 159 99 2 210 6 233 84 18 218 2 12 108 184 186 173 193 15 60 178 255 53", " "),
		},
		BeefyNextAuthoritySet: ibcgptypes.BeefyAuthoritySet{
			Id:   1523,
			Len:  5,
			Root: conv.SplitStrToSlice[byte]("48 72 3 250 90 145 217 133 44 170 254 4 180 184 103 164 237 39 160 122 91 238 61 21 7 180 177 135 166 135 119 162", " "),
		},
		ParachainHeads: conv.SplitStrToSlice[byte]("0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0", " "),
	},
}
var mmrProofs = ibcgptypes.MMRBatchProof{
	LeafIndexes: []uint64{15228},
	LeafCount:   15230,
	Items: [][]byte{
		conv.SplitStrToSlice[byte]("56 90 28 72 230 150 31 137 115 255 104 138 22 188 17 47 104 118 80 161 144 223 169 244 11 142 134 81 143 58 90 33", " "),
		conv.SplitStrToSlice[byte]("200 28 154 31 144 202 120 123 231 225 243 217 148 28 243 150 235 200 1 200 239 21 22 64 237 28 82 133 156 34 222 72", " "),
		conv.SplitStrToSlice[byte]("237 214 111 66 243 187 148 166 241 230 240 22 216 79 169 84 75 166 245 107 8 165 220 18 108 155 147 98 173 183 51 174", " "),
		conv.SplitStrToSlice[byte]("28 209 246 67 202 162 138 192 49 25 152 116 74 190 172 46 75 134 139 72 183 1 42 93 126 14 21 243 19 141 129 185", " "),
		conv.SplitStrToSlice[byte]("246 243 159 252 20 207 102 59 178 243 165 0 5 12 46 101 181 61 149 178 158 116 81 48 202 32 5 119 95 42 9 25", " "),
		conv.SplitStrToSlice[byte]("9 198 104 156 151 127 238 206 70 175 153 153 131 85 161 194 109 127 195 25 86 177 62 185 138 79 110 71 51 116 197 65", " "),
		conv.SplitStrToSlice[byte]("19 16 71 136 118 196 180 18 165 48 18 123 167 187 250 103 139 255 21 65 145 250 98 16 232 232 214 58 193 215 24 7", " "),
		conv.SplitStrToSlice[byte]("155 97 98 121 158 113 235 126 148 130 96 141 120 24 51 215 31 17 114 181 54 11 252 180 195 160 201 190 182 87 252 85", " "),
		conv.SplitStrToSlice[byte]("157 59 195 196 54 76 77 232 21 237 205 150 179 216 63 179 54 24 129 238 108 196 175 237 95 172 166 137 134 13 212 114", " "),
		conv.SplitStrToSlice[byte]("62 246 196 121 42 172 203 63 147 33 243 247 216 154 157 18 94 215 252 237 155 201 46 7 59 3 173 232 89 227 216 215", " "),
		conv.SplitStrToSlice[byte]("96 71 210 112 235 33 209 82 103 242 100 98 194 167 245 176 192 81 230 43 35 182 83 15 207 160 231 247 186 13 178 217", " "),
	},
}

var timestamp = ibcgptypes.StateProof{
	Key:   conv.SplitStrToSlice[byte]("240 195 101 195 207 89 214 113 235 114 218 14 122 65 19 196 159 31 5 21 244 98 205 207 132 224 241 214 4 93 252 187", " "),
	Value: conv.SplitStrToSlice[byte]("65 226 27 223 134 1 0 0", " "),
	Proofs: [][]byte{
		conv.SplitStrToSlice[byte]("128 37 0 128 142 246 240 95 231 69 116 43 210 254 241 34 99 155 97 211 254 143 92 9 20 101 35 51 93 83 189 13 67 56 35 90 128 220 70 71 23 254 52 56 235 177 36 105 170 135 170 194 216 76 175 82 236 55 36 76 104 67 131 83 248 73 88 183 80 128 178 63 109 241 206 119 144 131 77 148 131 194 21 61 11 185 116 14 138 104 76 114 131 163 2 139 25 123 204 189 137 97", " "),
		conv.SplitStrToSlice[byte]("128 255 255 128 43 105 28 49 22 122 76 34 116 19 197 229 33 220 109 27 50 217 90 1 106 49 78 208 47 24 183 8 226 191 227 194 128 88 165 100 195 193 243 95 216 51 190 128 175 138 115 151 235 0 78 2 153 89 98 239 6 24 116 186 61 214 79 11 96 128 178 138 163 191 45 255 173 145 31 163 79 167 224 33 78 102 215 32 123 116 229 162 203 139 184 32 242 79 236 32 203 80 128 141 227 167 160 95 29 233 224 192 244 84 27 195 129 26 26 7 23 222 68 139 2 36 24 185 47 48 127 139 17 139 37 128 170 117 218 169 102 122 107 159 196 34 166 212 144 230 172 178 82 49 85 255 217 125 205 180 182 38 172 230 57 39 155 233 128 56 235 236 115 168 46 39 37 182 175 90 228 27 202 122 183 255 35 152 53 128 146 178 48 227 199 3 241 168 76 101 32 128 86 194 186 63 24 63 16 119 157 61 106 246 16 188 145 237 70 120 216 186 17 153 11 163 182 227 165 208 109 91 190 214 128 91 104 33 50 197 41 8 112 85 38 5 127 115 171 127 204 171 74 246 215 42 152 5 99 77 216 211 204 83 241 48 209 128 194 212 77 55 30 95 193 245 2 39 215 73 26 214 90 208 73 99 3 97 206 251 74 177 132 72 49 35 118 9 240 131 128 209 39 41 214 88 255 20 228 60 32 73 23 161 116 213 221 178 84 230 28 53 255 237 187 134 162 96 180 215 51 144 174 128 127 130 71 70 31 59 67 239 23 189 136 106 130 43 160 53 112 229 213 47 13 254 224 0 99 180 72 30 52 208 43 92 128 11 58 50 2 144 211 159 129 123 108 240 85 115 125 62 6 118 94 226 232 129 50 51 139 18 19 128 127 40 236 79 168 128 203 8 254 178 80 245 195 97 81 30 20 246 8 95 3 55 158 125 141 228 188 35 201 117 31 191 179 177 228 129 213 83 128 151 219 231 206 232 208 227 242 29 136 101 102 234 31 189 100 63 195 84 49 30 183 211 52 219 97 153 19 180 151 33 103 128 107 253 187 240 224 190 220 185 147 182 92 156 234 30 146 154 86 215 138 59 123 197 61 27 124 166 252 72 142 34 149 238 128 154 107 212 113 167 39 139 108 163 125 58 34 103 26 92 150 231 20 15 118 14 157 70 16 160 81 108 48 186 179 115 62", " "),
		conv.SplitStrToSlice[byte]("158 195 101 195 207 89 214 113 235 114 218 14 122 65 19 196 16 2 80 95 14 123 144 18 9 107 65 196 235 58 175 148 127 110 164 41 8 0 0 104 95 15 31 5 21 244 98 205 207 132 224 241 214 4 93 252 187 32 65 226 27 223 134 1 0 0", " "),
	},
}
var subchainHeaders = ibcgptypes.SubchainHeaders{
	SubchainHeaders: []ibcgptypes.SubchainHeader{
		{
			BlockHeader: conv.SplitStrToSlice[byte]("47 216 3 114 152 227 5 216 13 102 171 177 171 42 162 230 31 2 25 186 207 149 240 199 194 194 210 186 242 107 103 28 241 237 30 204 179 138 195 242 182 180 186 147 36 13 228 97 228 74 112 192 94 94 127 118 118 49 100 85 199 28 60 139 210 148 117 230 243 112 13 148 23 147 186 26 152 150 124 142 223 228 31 214 58 104 70 86 13 90 162 149 6 80 224 28 130 100 12 6 66 65 66 69 181 1 3 3 0 0 0 252 91 173 16 0 0 0 0 90 151 0 59 38 211 212 185 236 250 0 124 120 195 207 220 59 215 164 225 233 118 11 232 211 18 102 224 116 139 214 57 194 134 221 169 197 103 217 7 43 199 12 143 193 212 8 192 34 91 14 31 215 42 251 199 25 178 203 73 30 210 173 4 215 134 129 148 72 114 241 183 227 191 90 88 209 182 31 194 187 200 69 38 10 255 78 123 123 209 138 92 167 79 109 8 4 66 69 69 70 132 3 131 79 104 195 33 161 208 242 156 164 3 120 80 122 102 198 67 105 240 96 40 47 16 197 136 94 190 101 145 9 176 52 5 66 65 66 69 1 1 208 127 127 118 216 8 235 77 108 81 233 99 32 222 215 228 227 208 13 148 245 78 91 212 178 88 123 136 179 19 147 36 214 167 135 71 164 59 192 52 166 67 91 55 227 152 52 110 99 71 64 140 68 160 88 108 54 6 174 19 69 139 58 130", " "),
			Timestamp:   timestamp,
		},
	},
}

var beefyMMR = ibcgptypes.BeefyMMR{
	SignedCommitment: signedCommitment,
	SignatureProofs:  signatureProofs,
	MmrLeavesAndBatchProof: ibcgptypes.MMRLeavesAndBatchProof{
		Leaves:        leaves,
		MmrBatchProof: mmrProofs,
	},
}

var gpHeader = ibcgptypes.Header{
	BeefyMmr: &beefyMMR,
	Message: &ibcgptypes.Header_SubchainHeaders{
		SubchainHeaders: &subchainHeaders,
	},
}

// timestamp unint64 value: 1678780392001
// timestamp string :       “2023-03-14 15:53:12.001 +0800 CST"
var consensusState = ibcgptypes.ConsensusState{
	Timestamp: time.UnixMilli(1678780392001),
	Root:      conv.SplitStrToSlice[byte]("30 204 179 138 195 242 182 180 186 147 36 13 228 97 228 74 112 192 94 94 127 118 118 49 100 85 199 28 60 139 210 148", " "),
}

func (suite *GrandpaTestSuite) TestMsgClientState() {
	// suite.Suite.T().Skip()
	// test pb marshal and unmarshal
	marshalCS, err := gpClientState.Marshal()
	// require.NoError(suite., err)
	suite.NoError(err)
	suite.Suite.T().Logf("marshal client state: %+v", marshalCS)
	// unmarshal
	// err = clientState.Unmarshal(marshalCS)
	var unmarshalCS ibcgptypes.ClientState
	err = unmarshalCS.Unmarshal(marshalCS)
	suite.NoError(err)
	suite.Equal(gpClientState, unmarshalCS)
	suite.Suite.T().Logf("unmarshal client state: %+v", unmarshalCS)
}

func (suite *GrandpaTestSuite) TestMsgConsensuseState() {
	// suite.Suite.T().Skip()
	// test pb marshal and unmarshal
	marshalCS, err := consensusState.Marshal()
	suite.NoError(err)
	suite.Suite.T().Logf("marshal consensusState: %+v", marshalCS)
	// unmarshal
	var unmarshalCS ibcgptypes.ConsensusState
	err = unmarshalCS.Unmarshal(marshalCS)
	suite.NoError(err)
	suite.Suite.T().Logf("unmarshal consensusState: %+v", unmarshalCS)
	suite.Suite.T().Logf("raw consensusState: %+v", consensusState)
	// Note: consensusState != unmarshalCS,because timestamp format is different!
	suite.NotEqual(consensusState, unmarshalCS, "timestamp is different,consensusState.timestamp: %+v,unmarshalCS.timestamp: %+v",
		consensusState.Timestamp, unmarshalCS.Timestamp)

	suite.Suite.T().Logf("unmarshal consensusState root: %+v", unmarshalCS.Root)
	suite.Suite.T().Logf("raw consensusState root: %+v", consensusState.Root)
	suite.Equal(consensusState.Root, unmarshalCS.Root)
	suite.Suite.T().Logf("unmarshal consensusState timestamp: %+v", unmarshalCS.Timestamp.UnixMilli())
	suite.Suite.T().Logf("raw consensusState timestamp: %+v", consensusState.Timestamp.UnixMilli())
	suite.Equal(consensusState.Timestamp.UnixMilli(), unmarshalCS.Timestamp.UnixMilli())

}

func (suite *GrandpaTestSuite) TestMsgBeefyMMR() {
	// suite.Suite.T().Skip()
	// test pb marshal and unmarshal
	marshalbeefyMMR, err := beefyMMR.Marshal()
	// require.NoError(suite., err)
	suite.NoError(err)
	suite.Suite.T().Logf("marshal beefy mmr: %+v", marshalbeefyMMR)
	// unmarshal
	// err = clientState.Unmarshal(marshalCS)
	var unmarshalBeefyMMR ibcgptypes.BeefyMMR
	err = unmarshalBeefyMMR.Unmarshal(marshalbeefyMMR)
	suite.NoError(err)
	suite.Suite.T().Logf("unmarshal beefy mmr: %+v", unmarshalBeefyMMR)
	suite.Equal(beefyMMR, unmarshalBeefyMMR)
}

func (suite *GrandpaTestSuite) TestMsgHeader() {
	// suite.Suite.T().Skip()
	// test pb marshal and unmarshal
	marshalGPHeader, err := gpHeader.Marshal()
	// require.NoError(suite., err)
	suite.NoError(err)
	suite.Suite.T().Logf("marshal grandpa pb header : %+v", marshalGPHeader)
	// unmarshal
	// err = clientState.Unmarshal(marshalCS)
	var unmarshalGPHeader ibcgptypes.Header
	err = unmarshalGPHeader.Unmarshal(marshalGPHeader)
	suite.NoError(err)
	suite.Suite.T().Logf("unmarshal grandpa pb header: %+v", unmarshalGPHeader)
	suite.Equal(gpHeader, unmarshalGPHeader)
}

func (suite *GrandpaTestSuite) TestDecodeStateProof() {
	// var stateProof beefy.StateProof
	var stateProof beefy.StateProof
	proof := conv.SplitStrToSlice[byte]("244 101 204 242 3 105 192 221 218 216 45 16 3 82 58 196 142 47 99 99 80 87 249 175 51 232 228 192 128 156 156 52 197 28 84 188 153 203 33 207 91 213 249 92 133 178 76 32 224 48 99 111 110 110 101 99 116 105 111 110 45 51 69 1 1 60 48 55 45 116 101 110 100 101 114 109 105 110 116 45 51 48 49 48 45 103 114 97 110 100 112 97 45 48 0 12 105 98 99 4 4 49 8 52 79 82 68 69 82 95 79 82 68 69 82 69 68 60 79 82 68 69 82 95 85 78 79 82 68 69 82 69 68 0 0 0 0 0 0 0 0 0 0 0 0 24 69 1 1 60 48 55 45 116 101 110 100 101 114 109 105 110 116 45 51 48 49 48 45 103 114 97 110 100 112 97 45 48 0 12 105 98 99 4 4 49 8 52 79 82 68 69 82 95 79 82 68 69 82 69 68 60 79 82 68 69 82 95 85 78 79 82 68 69 82 69 68 0 0 0 0 0 0 0 0 0 0 0 0 252 63 26 12 84 188 153 203 33 207 91 213 249 92 133 178 76 32 224 48 99 111 110 110 101 99 116 105 111 110 45 51 244 122 31 9 126 87 41 102 33 103 173 147 84 209 126 255 32 246 157 107 81 25 249 34 231 142 228 165 40 183 106 235 161 2 128 168 20 128 55 92 36 190 155 31 216 140 77 72 159 87 116 23 200 119 47 164 112 211 69 65 75 124 130 170 193 146 97 120 235 199 128 173 88 81 204 161 200 86 27 223 14 49 224 201 113 96 61 83 123 17 1 119 135 13 53 54 106 62 231 29 91 190 169 128 167 80 60 9 242 55 155 253 158 34 247 241 17 176 253 1 101 249 172 79 37 152 61 114 185 90 140 224 83 75 128 220 128 7 252 233 77 152 160 180 150 177 174 169 91 252 121 210 32 38 62 58 92 30 176 163 221 178 228 89 55 7 229 18 206 128 81 119 54 240 89 55 154 228 116 2 207 155 93 147 6 215 50 179 165 115 98 242 54 144 218 23 76 154 206 45 22 132 77 8 128 255 255 128 221 224 25 195 37 50 35 29 233 146 57 141 55 66 69 69 191 56 96 150 98 44 12 6 28 162 198 120 143 17 218 82 128 7 26 227 143 71 73 167 232 156 37 93 55 246 49 224 224 216 221 185 123 32 53 76 224 7 233 156 67 186 224 157 58 128 156 48 35 172 194 185 99 207 129 3 147 83 236 156 131 252 8 213 7 199 88 184 76 140 7 183 162 169 166 108 57 163 128 232 223 188 97 167 118 135 17 204 20 34 164 222 249 49 205 192 81 141 212 185 160 84 134 192 246 35 76 222 212 181 72 128 3 55 242 53 158 236 167 65 87 77 185 28 246 118 243 245 154 142 154 59 236 90 142 93 209 147 183 62 164 138 133 212 128 60 108 53 1 162 122 200 54 122 81 19 58 17 119 21 17 235 31 209 248 224 119 115 89 114 164 96 95 222 254 204 10 128 56 56 7 159 108 125 92 223 222 248 56 191 36 5 189 54 6 42 210 197 112 90 246 138 76 198 150 246 121 181 138 236 128 91 104 33 50 197 41 8 112 85 38 5 127 115 171 127 204 171 74 246 215 42 152 5 99 77 216 211 204 83 241 48 209 128 194 212 77 55 30 95 193 245 2 39 215 73 26 214 90 208 73 99 3 97 206 251 74 177 132 72 49 35 118 9 240 131 128 48 114 47 191 154 198 168 106 153 186 232 246 156 74 160 145 38 80 140 177 219 228 43 119 73 243 108 48 70 16 44 228 128 177 12 58 95 62 173 184 43 139 216 194 238 76 232 183 111 207 194 58 32 94 189 21 0 20 221 175 195 172 124 118 213 128 94 171 175 160 186 230 240 205 172 249 218 142 143 39 67 137 171 136 200 168 224 32 66 97 231 24 226 116 203 244 233 68 128 64 139 173 24 111 31 197 24 172 162 254 120 107 221 231 4 113 14 54 92 14 94 149 167 205 132 196 144 113 184 62 241 128 63 30 43 134 227 69 144 147 249 112 150 86 121 71 129 176 242 209 197 118 15 22 61 28 213 107 153 139 146 41 60 175 128 107 253 187 240 224 190 220 185 147 182 92 156 234 30 146 154 86 215 138 59 123 197 61 27 124 166 252 72 142 34 149 238 128 79 21 231 231 140 63 206 72 168 55 190 124 80 15 131 43 128 241 173 202 93 192 86 78 249 168 83 121 44 147 154 161 141 4 158 204 242 3 105 192 221 218 216 45 16 3 82 58 196 142 189 38 104 95 10 188 245 27 222 65 138 14 86 107 122 75 66 31 59 242 32 4 0 0 0 0 0 0 0 128 117 175 59 227 121 139 132 88 210 154 21 108 159 120 130 68 142 171 42 19 161 225 225 48 217 137 130 110 211 201 141 120 128 231 89 48 111 230 84 10 136 36 51 167 106 251 133 149 228 66 208 23 165 178 29 2 155 163 16 92 205 135 69 225 79 80 95 14 123 144 18 9 107 65 196 235 58 175 148 127 110 164 41 8 0 0 128 167 44 230 196 161 193 83 136 254 11 203 195 47 140 40 126 147 20 197 59 238 224 14 71 56 215 194 215 113 230 93 4 104 95 13 175 218 65 33 225 150 51 237 160 123 37 248 10 100 93 32 4 0 0 0 0 0 0 0 128 116 162 77 184 92 31 93 203 250 126 72 89 162 229 185 37 84 45 136 23 179 38 116 8 122 15 23 34 219 32 146 120 128 227 115 137 30 20 181 203 2 254 169 29 131 190 156 17 33 64 22 73 143 67 47 186 163 37 26 72 188 138 193 93 46 128 25 169 67 217 185 106 209 32 253 239 179 234 10 209 111 219 224 175 21 25 252 157 196 28 89 243 55 112 98 101 171 67 217 1 159 15 99 99 80 87 249 175 51 232 228 192 128 156 156 52 197 42 0 128 55 8 119 5 205 216 220 255 4 14 197 93 38 163 170 78 208 60 162 58 141 180 174 118 202 200 10 80 96 156 216 224 128 8 166 74 27 62 132 74 244 93 162 110 73 187 158 94 155 23 74 125 54 81 249 161 203 81 135 148 205 159 188 231 237 128 162 143 219 159 224 125 250 143 9 176 124 233 214 236 96 217 238 12 105 137 122 248 18 228 118 114 183 178 126 140 199 86", " ")
	err := gsrpccodec.Decode(proof, &stateProof)
	// err = cdc.Unmarshal(existenceProof.Value, &stateProof)
	if err != nil {
		suite.Suite.T().Logf("🐙🐙 ics10::TestDecodeStateProof -> decoded stateProof err:%+v ", err)

	}
	suite.Suite.T().Logf("🐙🐙 ics10::produceVerificationArgs -> decoded stateProof:%+v ", stateProof)
	stateRoot := conv.SplitStrToSlice[byte]("155 161 120 40 55 234 101 167 69 91 161 30 249 87 16 29 196 126 34 124 193 228 232 16 73 36 69 248 61 62 226 50", " ")
	// err = beefy.VerifyStateProof(stateProof.Proofs, consensusState.Root, stateProof.Key, stateProof.Value)
	err = beefy.VerifyStateProof(stateProof.Proofs, stateRoot, stateProof.Key, stateProof.Value)
	if err != nil {
		suite.Suite.T().Logf("🐙🐙 ics10::TestDecodeStateProof -> beefy.VerifyStateProof err:%+v ", err)

	}

}
