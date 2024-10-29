package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type StudioChaincode struct {
}

type Material struct {
	Name     string  `json:"name"`
	Origin   string  `json:"origin"`
	Quantity int64   `json:"quantity"`
	ObjectType string `json:"objectType"`
}

type Wand struct {
	Name       string     `json:"name"`
	ProductionHistory []Production   `json:"productionHistory"`
	Quantity   int64      `json:"quantity"`
	ObjectType string `json:"objectType"`
}

type Production struct {
	Materials []Material `json:"materials"` // Materiais utilizados na produção
	Available   bool      `json:"bool"` // Disponivel para venda ou nao
}

// Init inicializa o chaincode
func (t *StudioChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke é a entrada para invocar transações no chaincode
func (t *StudioChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "addMaterial" {
		return t.addMaterial(stub, args)
	} else if function == "listMaterials" {
		return t.listMaterials(stub)
	} else if function == "createWand" {
		return t.createWand(stub, args)
	} else if function == "sellWand" {
		return t.sellWand(stub, args)
	} else if function == "listWands" {
		return t.listWands(stub)
	}

	return shim.Error("Funcao invalida.")
}
// Função para adicionar material
func (t *StudioChaincode) addMaterial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Verifica se ha 3 exatos argumentos
	if len(args) != 3 {
		return shim.Error("Esperado 3 argumentos: Nome, Origem, Quantidade.")
	}
	// Verifica se foi fornecido uma quantidade valida
	quantity, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error("Quantidade inválida.")
	}
	// Adiciona quantidade em cada material
	material, err := t.getMaterial(stub, args[0], args[1])
	if err != nil {
		return shim.Error("Erro ao obter material.")
	}
	material.Quantity += quantity
	// Serializa material
	materialBytes, err := json.Marshal(material)
	if err != nil {
		return shim.Error("Erro ao serializar material.")
	}
	// Adiciona a ledger
	err = stub.PutState(args[0]+args[1], materialBytes)
	if err != nil {
		return shim.Error("Erro ao salvar material.")
	}
	return shim.Success(nil)
}
// Função para listar todos os materiais cadastrados no ledger
func (t *StudioChaincode) listMaterials(stub shim.ChaincodeStubInterface) pb.Response {
	// Obtém um iterador para todos os registros no ledger, usando um intervalo vazio para capturar tudo
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error("Erro ao obter materiais.")
	}
	// Garante que o iterador seja fechado após o uso para liberar recursos
	defer resultsIterator.Close()
	// Cria um slice para armazenar todos os materiais encontrados
	var materials []Material
	// Itera por todos os registros no ledger
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Erro ao iterar por materiais.")
		}
		var material Material
		// Deserializa o valor JSON armazenado para a estrutura Material
		err = json.Unmarshal(queryResponse.Value, &material)
		// Verifica se a deserialização ocorreu sem erros e se o tipo do objeto é "Material"
		if err == nil && material.ObjectType == "Material" {
			materials = append(materials, material)
		}
	}
	// Serializa a lista de materiais para JSON
	materialsBytes, err := json.Marshal(materials)
	if err != nil {
		// Retorna um erro caso a serialização falhe
		return shim.Error("Erro ao serializar materiais.")
	}
	// Retorna com sucesso a lista de materiais em formato JSON
	return shim.Success(materialsBytes)
}

// Função para criar uma varinha ou incrementar a quantidade se ela já existir
func (t *StudioChaincode) createWand(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Verifica a organização que está executando a transação
	orgID, err := getOrg(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("Erro ao obter organização: %s", err))
	}
	// Verifica se a organização é permitida para criar varinhas
	if orgID != "org0-example-com" {
		return shim.Error(fmt.Sprintf("Somente o Sr. Olivaras pode criar novas varinhas. Sua organização: %s.", orgID))
	}
    // Verifica se há pelo menos 4 argumentos
    if len(args) < 4 {
        return shim.Error("Esperado pelo menos 4 argumentos: Nome da Varinha, Nome Material, Origem do Material, Quantidade do material.")
    }
    // Verifica se os argumentos dos materiais foram passados corretamente
    if len(args) % 3 != 1 || len(args) == 1{
        return shim.Error("Argumentos dos materiais devem ser em trios (Nome do Material, Origem, Quantidade).")
    }
	// Cria o nome da varinha e um vetor materiais
    wandName := args[0]
    var materials []Material
    // Primeira verificação: garante que todos os materiais tenham quantidade suficiente antes de qualquer atualização
    for i := 1; i < len(args)-2; i += 3 {
        materialName := args[i]
        origin := args[i+1]
        quantityRequired, err := strconv.ParseInt(args[i+2], 10, 64)
        if err != nil {
            return shim.Error("Erro ao converter quantidade do material.")
        }
        materialBytes, err := stub.GetState(materialName + origin)
        if err != nil {
            return shim.Error("Erro ao obter material.")
        }
        if materialBytes == nil {
            return shim.Error(fmt.Sprintf("Material %s de origem %s não encontrado.", materialName, origin))
        }
        var material Material
        err = json.Unmarshal(materialBytes, &material)
        if err != nil {
            return shim.Error("Erro ao deserializar material.")
        }
        if material.Quantity < quantityRequired {
            return shim.Error(fmt.Sprintf("Não há quantidade suficiente do material %s. Quantidade disponível: %d, requerida: %d.", material.Name, material.Quantity, quantityRequired))
        }
        // Armazena temporariamente os materiais que passaram na validação
        materials = append(materials, Material{
            Origin:    origin,
            Name:      material.Name,
            Quantity:  quantityRequired,
			ObjectType: "Material",
        })
    }

	// Segunda etapa: realiza a criação ou incremento da varinha e desconta os materiais
	wandBytes, err := stub.GetState(wandName)
	if err != nil {
		return shim.Error("Erro ao verificar existência da varinha.")
	}
	var wand Wand
	if wandBytes != nil {
    // Caso a varinha já exista, incrementa a quantidade
    err = json.Unmarshal(wandBytes, &wand)
    if err != nil {
        return shim.Error("Erro ao desserializar varinha existente.")
    }
    wand.Quantity += 1
    // Adiciona os materiais utilizados na produção ao histórico de produções
    production := Production{
        Materials: materials,
		Available: true,
    }
    wand.ProductionHistory = append(wand.ProductionHistory, production)
	} else {
    // Caso a varinha não exista, cria uma nova com quantidade 1 e adiciona o histórico
		wand = Wand{
			Name:             wandName,
			Quantity:         1,
			ProductionHistory: []Production{{Materials: materials, Available: true}},
			ObjectType:       "Wand",
		}
	}
    // Atualiza as quantidades de cada material no ledger após validação completa
    for _, material := range materials {
        materialKey := material.Name + material.Origin
        materialBytes, err := stub.GetState(materialKey)
        if err != nil {
            return shim.Error("Erro ao obter material para atualizar quantidade.")
        }
		// Tenta desserializar
        var ledgerMaterial Material
        err = json.Unmarshal(materialBytes, &ledgerMaterial)
        if err != nil {
            return shim.Error("Erro ao deserializar material para atualização.")
        }
        // Desconta a quantidade usada
        ledgerMaterial.Quantity -= material.Quantity
        materialBytes, err = json.Marshal(ledgerMaterial)
        if err != nil {
            return shim.Error("Erro ao serializar material atualizado.")
        }
		// Atualiza a quantidade de material na ledger
        err = stub.PutState(materialKey, materialBytes)
        if err != nil {
            return shim.Error("Erro ao atualizar material no ledger.")
        }
    }
    // Serializa a varinha para salvar no ledger
    wandBytes, err = json.Marshal(wand)
    if err != nil {
        return shim.Error("Erro ao serializar varinha.")
    }
    // Salva a varinha no ledger
    err = stub.PutState(wand.Name, wandBytes)
    if err != nil {
        return shim.Error("Erro ao salvar varinha.")
    }

    return shim.Success([]byte(fmt.Sprintf("Varinha '%s' criada ou atualizada com sucesso.", wand.Name)))
}

// sellWand vende uma varinha, identificada pelo nome e posição no vetor ProductionHistory
func (t *StudioChaincode) sellWand(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// Verifica a organização que está executando a transação
	orgID, err := getOrg(stub)
	if err != nil {
		return shim.Error("Erro ao obter organização: " + err.Error())
	}
	// Verifica se a organização é permitida para vender varinhas
	if orgID != "org0-example-com" {
		return shim.Error(fmt.Sprintf("Somente o Sr. Olivaras pode vender varinhas. Sua organização: %s", orgID))
	}
	// Verifica os argumentos passados
    if len(args) != 2 {
        return shim.Error("Esperado 2 argumentos: Nome da varinha e índice no vetor ProductionHistory.")
    }
    // Passo 1: Verificar se a varinha com aquele nome existe na ledger
    wandName := args[0]
    wandBytes, err := stub.GetState(wandName)
    if err != nil || wandBytes == nil {
        return shim.Error(fmt.Sprintf("Varinha com o nome '%s' não encontrada na ledger.", wandName))
    }
    // Passo 2: Verificar se o segundo argumento (índice) é um número inteiro
    i, err := strconv.Atoi(args[1])
    if err != nil {
        return shim.Error("O índice deve ser um número inteiro.")
    }
    // Deserializar a varinha
    var wand Wand
    err = json.Unmarshal(wandBytes, &wand)
    if err != nil {
        return shim.Error("Erro ao decodificar dados da varinha.")
    }
    // Passo 3: Verificar se o índice está dentro do limite do vetor ProductionHistory
    if i < 0 || i >= len(wand.ProductionHistory) {
        return shim.Error("Índice fora do limite do vetor ProductionHistory.")
    }
    // Alterar o atributo Available do item no índice especificado para false se a varinha estiver disponivel
    if wand.ProductionHistory[i].Available{
		wand.Quantity -= 1
		wand.ProductionHistory[i].Available = false
	} else {
		return shim.Error("Essa varinha ja foi vendida.")
	}
    // Passo 4: Gravar a varinha de volta na ledger
    updatedWandBytes, err := json.Marshal(wand)
    if err != nil {
        return shim.Error("Erro ao codificar a varinha atualizada.")
    }
    err = stub.PutState(wandName, updatedWandBytes)
    if err != nil {
        return shim.Error("Erro ao atualizar a varinha na ledger")
    }
    return shim.Success([]byte("Varinha vendida com sucesso"))
}
// Função para listar todas as varinhas disponiveis/indisponiveis
func (t *StudioChaincode) listWands(stub shim.ChaincodeStubInterface) pb.Response {
	// Obtém um iterador para todos os registros no ledger, usando um intervalo vazio para capturar tudo
	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error("Erro ao obter varinhas.")
	}
	// Garante que o iterador seja fechado após o uso para liberar recursos
	defer resultsIterator.Close()
	// Cria um slice para armazenar todos os materiais encontrados
	var wands []Wand
	// Itera por todos os registros no ledger
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Erro ao iterar por varinhas.")
		}
		var wand Wand
		// Deserializa o valor JSON armazenado para a estrutura Wand
		err = json.Unmarshal(queryResponse.Value, &wand)
		// Verifica se a deserialização ocorreu sem erros e se o tipo do objeto é "Wand"
		if err == nil && wand.ObjectType == "Wand" {
			wands = append(wands, wand)
		}
	}
	// Serializa a lista de materiais para JSON
	wandsBytes, err := json.Marshal(wands)
	if err != nil {
		// Retorna um erro caso a serialização falhe
		return shim.Error("Erro ao serializar varinhas.")
	}
	// Retorna com sucesso a lista de materiais em formato JSON
	return shim.Success(wandsBytes)
}
// GetMaterial retorna um material pelo materialName e Origem
func (t *StudioChaincode) getMaterial(stub shim.ChaincodeStubInterface, materialName string, origin string) (*Material, error) {
    // Verificar se o materialName ou origin estão vazios
    if materialName == "" || origin == "" {
        return nil, fmt.Errorf("nome do Material ou origem não podem ser vazios")
    }
    // Concatena o Nome do Material e a origem para formar a chave composta
    compositeKey := materialName + origin
    // Busca o estado (material) pela chave composta
    materialJSON, err := stub.GetState(compositeKey)
    if err != nil {
        return nil, fmt.Errorf("erro ao buscar estado para o material %s: %v", compositeKey, err)
    }
    // Se o material não for encontrado, retorna um novo objeto vazio com os valores fornecidos
    if materialJSON == nil {
        material := &Material{
            Name:     materialName,
            Quantity: 0,
            Origin:   origin,
			ObjectType: "Material",
        }
        return material, nil
    }
    // Deserializar os dados do JSON para o struct Material
    var material Material
    err = json.Unmarshal(materialJSON, &material)
    if err != nil {
        return nil, fmt.Errorf("erro ao desserializar material: %s", err)
    }
    // Retornar o material encontrado
    return &material, nil
}
// getOrg é uma função que obtém o identificador da organização (MSP ID) do criador da transação.
func getOrg(stub shim.ChaincodeStubInterface) (string, error) {
	// Obtém o criador da transação a partir do stub.
	creator, err := stub.GetCreator()
	if err != nil {
		return "", fmt.Errorf("erro ao ler o criador da transacao: %v", err)
	}
	// Extrai o MSP ID do criador da transação.
	mspID, err := getMSPID(creator)
	if err != nil {
		return "", err
	}
	return mspID, nil
}
// getMSPID é uma função auxiliar que desserializa os dados do criador da transação para obter
// o MSP ID (identificador da organização) associado a ele. 
func getMSPID(creator []byte) (string, error) {
	// Cria uma estrutura de identidade serializada para armazenar o criador desserializado.
	identity := &msp.SerializedIdentity{}
	// Desserializa o criador usando o protocolo proto.
	if err := proto.Unmarshal(creator, identity); err != nil {
		return "", fmt.Errorf("erro ao desserializar o criador: %v", err)
	}
	// Retorna o MSP ID extraído da identidade desserializada.
	return identity.Mspid, nil
}
func main() {
	err := shim.Start(new(StudioChaincode))
	if err != nil {
		fmt.Printf("Erro ao iniciar chaincode: %s.", err)
	}
}