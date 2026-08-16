# GottPolling

* [Objetivo](#objetivo)
* [Tecnologias](#tecnologias)

# sync/offline-first

  O GottPolling foi desenvolvido na intenção de conectar um servidor online
a um servidor local sem precisar lhe dá com a camada de rede NAT, permitindo
a comunicação sem aberturas de portas no roteador, mas sim atravéz de polling,
sendo esse metodo adotado em vez de websocket.

# Objetivo
## O objetivo do projeto é não precisar pagar um banco de dados...
  Buscando uma solução para poder conectar dois servidores sendo um local e outro em uma VPS
de baixo custo ou grátis, o GottPolling propõe um solução para manter renderizando na aplicação
em uma VPS grátis os dados efemeros de um banco de dados SQLite, que podem ser apagados após um
novo deploy ou tempo de inatividade que leva a instância a hibernar e apaga arquivos como databases
em SQLite ou Json do fileSystem. Como falado acima a intenção é não precisar pagar um banco de dados
caro logo de inicio (obviamente observando a nescessidade de cada projeto), permitindo escalar um pouco
mais ou até o nivel que a empresa preferir, notoriamente o SQLite na maquina local não é a única opção,
porém nessa camada de proposta inicial a idéia é poder renderizar os dados de forma leve na internet,
atravéz do SQLite.
  Nessa abordagem há dois bancos de dados sicronizando entre si, um em uma VPS e outro localmente, havendo
a exclusão de qualquer um ambos se restauram pois contém os mesmos dados e se comunicam entre si, por usar
a abordagem de offline-first.
* ** existe a nescessidade de um servidor local rodando porém com engenhosidade pode se encontrar algumas soluções interessantes que permitam desligar o servidor local, pois é essa uma das propostas do esquema apresentado.**
Não me deterei em falar sobre essas soluções.

# Tecnologias

* **Golang:**
* **SQLite:**




