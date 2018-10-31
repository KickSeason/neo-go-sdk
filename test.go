type NeoTxData struct {
	Vin []NeoTxIn `json:"vin"`
	Vout []NeoTxOut `json:"vout"`
}

type NeoTxIn struct {
	Txid         string `json:"txid"`
	Vout         uint16 `json:"vout"`
	Address      string `json:"address"`
}

func SignNeoTxWithPrivKey(txdata *NeoTxData, prikeys []string) (string, string, int) {

	utxos_in := make([]*neotransaction.UTXO, 0, len(txdata.Vin))
	utxos_out := make([]*neotransaction.UTXO, 0, len(txdata.Vout))
	for _, vin := range txdata.Vin {
		utxo := &neotransaction.UTXO{}
		utxo.TxHash, _ = hex.DecodeString(strings.TrimPrefix(vin.Txid, "0x"))
		utxo.TxHash = neoutils.Reverse(utxo.TxHash)
		utxo.Index = vin.Vout
		utxos_in = append(utxos_in, utxo)
	}
	for _, vout := range txdata.Vout {
		utxo := &neotransaction.UTXO{}
		assetid := getAssetId(vout.AssetSymbol)
		if assetid == "" {
			return "", "", errcode.INVALID_INPUT_DATA
		}
		utxo.AssetID, _ = hex.DecodeString(assetid)
		utxo.AssetID = neoutils.Reverse(utxo.AssetID)
		utxo.Value = vout.Value * neotransaction.TxOutputValueBase
		address, _ := neotransaction.ParseAddress(vout.Address)
		utxo.ScriptHash = address.ScripHash
		utxos_out = append(utxos_out, utxo)
	}

	tx := neotransaction.CreateContractTransaction()

	for _, utxo := range utxos_in {
		tx.AppendInput(utxo)
	}
	for _, utxo := range utxos_out {
		address, _ := neotransaction.ParseAddressHash(utxo.ScriptHash)
		tx.AppendOutput(address, utxo.AssetID, utxo.Value)
	}

	for _, key := range prikeys {
		keypair, _ := neotransaction.DecodeFromWif(key)
		tx.AppendAttribute(neotransaction.UsageScript, keypair.CreateBasicAddress().ScripHash)
		tx.AppendBasicSignWitness(keypair)
	}

	return tx.TXID(), tx.RawTransactionString(), errcode.OK
}

func getAssetId(assetSymbol string) string {
	switch assetSymbol {
	case "neo", "Neo", "NEO":
		return neotransaction.AssetNeoID
	case "gas", "Gas", "GAS":
		return neotransaction.AssetGasID
	default:
		return ""
	}
}