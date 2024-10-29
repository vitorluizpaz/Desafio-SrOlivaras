# Smart Contract Studio

Este projeto utiliza como base o [Minifabric](https://github.com/hyperledger-labs/minifabric), um ambiente para facilitar a configuração e execução de redes Hyperledger Fabric.

O objetivo deste projeto é a implementação de um smart contract, para fornecer uma interface para que bruxos produtores de matéria-prima possam registrar e gerenciar novos materiais no estoque, contribuindo para a cadeia de produção de varinhas mágicas. Isso permite que Olivaras consulte os materiais disponíveis e fabrique varinhas utilizando os itens cadastrados.

Além disso, o contrato garante a rastreabilidade de cada varinha produzida, registrando os materiais e suas respectivas origens (produtores) no sistema. Para reforçar a segurança, somente o Sr. Olivaras tem permissão para criar e vender varinhas, com a organização padrão Org0 do Minifabric representando Olivaras, enquanto a cadeia de produção de matéria-prima é gerida pela organização Org1.

# Como executar a solução no Windows 10 (Para executar em outro sistema há uma pequena mudança na forma como os argumentos são passados. Para verificar tais mudanças acessar a documentação do minifabric):

## Pré-requisitos
[docker](https://www.docker.com/) (18.03 or newer) environment

[Minifabric](https://github.com/hyperledger-labs/minifabric) environment

## Realizando o Deploy do Smart Contract
Abra o terminal em ~/mywork e utilize o comando:

minifab up

Dentro do diretório que foi instalado o Minifabric, navegar até o diretório vars/chaincode/, neste diretório salvar a pasta studio.

### Realizar o Deploy
Abra o terminal em ~/mywork para realizar o deploy

Parametro -n studio representa o nome do smart contract

Parametro -l go representa a linguagem do smart contract

Parametro -v 1.2 é a versão que está sendo realizada o deploy

```
minifab ccup -n studio -l go -v 1.2
```

## Métodos Disponíveis no Smart Contract
### Adicionar Material
Registra novos materiais no estoque:

Parametro -n studio

Parametro -p nome do método

Parametros obrigatórios: Nome do Material, Origem e Quantidade

```
minifab invoke -n studio -p \"addMaterial\",\"Carvalho\",\"Bruxo1\",\"10\"
```

### Consultar Materiais Disponíveis
Permite visualizar os materiais cadastrados e disponíveis:

Parametro -n studio

Parametro -p nome do método

```
minifab invoke -n studio -p \"listMaterials\"
```

### Criar Nova Varinha
Gera uma nova varinha mágica a partir dos materiais disponíveis:

Parametro -n studio

Parametro -p nome do método

Parametros obrigatórios: Nome da Varinha, Nome do Material, Origem e Quantidade (Pode-se adicionar mais materiais em trios - Nome, origem e quantidade)

```
minifab invoke -n studio -p \"createWand\",\"ElderWand\",\"Carvalho\",\"Bruxo1\",\"5\"
minifab invoke -n studio -p \"createWand\",\"ElderWand\",\"Carvalho\",\"Bruxo1\",\"5\","Flor","Bruxo2","10"
```

### Registrar Venda de Varinha
Vende a varinha

Parametro -n studio

Parametro -p nome do método

Parametros obrigatórios: Nome da Varinha, Indice relativo ao historico de producao.

```
minifab invoke -n studio -p \"sellWand\",\"ElderWand\",\"0\"
```
### Consultar Varas Disponíveis 
Lista as varinhas mágicas disponiveis (available = true) e indisponíveis
(available = false).

Parametro -n studio

Parametro -p nome do método

```
minifab invoke -n studio -p \"listWands\"
```