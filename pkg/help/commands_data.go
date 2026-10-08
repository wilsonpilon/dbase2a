package help

// CommandsSummaryText contém o catálogo completo e resumo de uma linha de cada comando do dBASE II++.
const CommandsSummaryText = `================================================================================
                    dBase II++ — RESUMO DE COMANDOS (COMMANDS)
================================================================================
  Comando        Descrição
  ------------   ---------------------------------------------------------------
  APPEND         Adiciona novos registros ao final do arquivo de dados ativo
  BROWSE         Navega e edita registros em modo visual interativo de tabela
  CLEAR / CLS    Limpa a tela do terminal e reposiciona o cursor no topo
  CONTINUE       Continua a busca sequencial iniciada pelo comando LOCATE
  COPY           Copia registros ou estrutura para um novo arquivo .DBF
  COUNT          Conta o número de registros que satisfazem uma condição
  CREATE         Cria a estrutura de um novo arquivo de banco de dados (.DBF)
  DELETE         Marca registro(s) como excluído(s) logicamente
  DISPLAY        Exibe registros, campos, expressões ou status na tela
  DO             Executa um arquivo de programa ou procedimento (.CMD / .PRG)
  EDIT           Exibe e permite a edição interativa de um registro específico
  FIND           Localiza rapidamente um registro em arquivo indexado por chave
  GOTO / GO      Posiciona o ponteiro de registro no número especificado
  HELP           Exibe documentação de ajuda interativa sobre comandos e sintaxe
  INDEX          Cria um arquivo de índice (.NDX) ordenado por campo ou expressão
  LIST           Lista registros ou estrutura de dados sem pausa automática
  LOCATE         Procura sequencialmente por um registro que atenda a uma condição
  MODIFY         Modifica a estrutura de um arquivo (.DBF) ou edita um programa
  PACK           Remove fisicamente todos os registros marcados para exclusão
  QUIT / EXIT    Encerra a sessão do dBase II++ e fecha todos os arquivos abertos
  RECALL         Restaura registro(s) marcado(s) para exclusão lógica
  REPLACE        Substitui o conteúdo de campos por novos valores ou expressões
  RESTORE        Restaura variáveis de memória salvas em arquivo .MEM
  SAVE           Salva as variáveis de memória ativas em arquivo .MEM
  SELECT         Alterna entre áreas de trabalho primária e secundária
  SET            Configura parâmetros, formatos e variáveis de ambiente
  SKIP           Desloca o ponteiro de registro para frente ou para trás
  SORT           Gera um novo arquivo ordenado por campos específicos
  STORE          Atribui o resultado de uma expressão a uma variável de memória
  SUM            Totaliza valores numéricos de campos com filtro condicional
  TOTAL          Gera arquivo sumarizado a partir de tabela classificada
  USE            Abre um arquivo de banco de dados (.DBF) e seus índices
  ZAP            Exclui instantaneamente todos os registros da tabela ativa
  ? / ??         Avalia e exibe o valor de uma expressão na tela
  @ SAY / GET    Posiciona e formata texto ou campos de entrada na tela
================================================================================`
