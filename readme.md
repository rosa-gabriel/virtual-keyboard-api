# Vitual Keyboard

**Desenvolvedores**: Gabriel Eduardo da Silva Rosa

Documentação do [FrontEnd](https://github.com/SirOtaiv/Security-Box)

## Como rodar?

É necessário ter o Go instalado na máquina, mais especificamente a versão `1.24.0`

O comando abaixo ira baixar as dependencias necessárias:
```go
go run .
```

Antes de rodar a aplicação é necessário subir um postgresql com que o a aplicação se conectara, caso você tenha um ambiente kubernetes disponivel é possivel subir o banco com o comando:
```
kubectl apply -f ./postgres.yml
```

> Uma maneira simples de instalar o kubernetes localmente é utilizando o `minikube`

## Como funciona?

### Combinações únicas

Para garantir as combinação únicas foram seedadas 100 combinações no arquivo `seed.go`

Esse arquivo vai criar todas as tabelas no banco caso não existam, inserir os usuários padrão e inserir as possibilidades no banco.

### Session validation

Toda vez que um fluxo de login se inicia é um hash que será validado no depara de senha do login, garantindo que os botões clicados foram gerados pelo servidor.

### Password Hashing

A senha é criptografada no banco utilizando o algoritmo BCrypt.

> Um problema na escolha desse algoritmo foi que a sua velocidade não é muito grande

### Comparação das senhas

No fluxo de validação da senha gerada todas as combinações possiveis dos digitos são comparadas com a senhas de todos os usuários, se alguma senha der match o usuário estara logado.

## Tecnologias

Golang -> Linguagem moderna de desenvolvimento web

Fuego -> Frameword golang responsável por gerar a spec Open API da API

Postgres -> Banco relacional Open-Source e padrão no mercado

## Dificuldades

A maior dificuldade foi garantir a segurança dentro dos escopos do projeto.
