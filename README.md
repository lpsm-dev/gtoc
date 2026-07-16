<!-- BEGIN_DOCS -->
<div align="center">

[🇺🇸 English Version](README_en.md)

<a name="readme-top"></a>

Hello Human 👽! Bem-vindo ao meu repositório 👋

<img src="https://github.com/lpsm-dev/lpsm-dev/blob/5cf57b68283a857a105430d0d6c9290ee657a012/.github/assets/go-cli.png" width="350"/>

**Gere uma bela documentação para os seus repositórios Git**

[![CI](https://github.com/lpsm-dev/gtoc/actions/workflows/ci.yaml/badge.svg)](https://github.com/lpsm-dev/gtoc/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lpsm-dev/gtoc)](https://goreportcard.com/report/github.com/lpsm-dev/gtoc)
[![Commitizen friendly](https://img.shields.io/badge/commitizen-friendly-brightgreen.svg)](https://www.conventionalcommits.org/en/v1.0.0/)
[![Semantic Release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://semantic-release.gitbook.io/semantic-release/usage/configuration)
[![Built with Devbox](https://jetpack.io/img/devbox/shield_galaxy.svg)](https://jetpack.io/devbox/docs/contributor-quickstart/)

📌 Curta esse repositório para acompanhar atualizações e novidades ( ≖‿ ≖ )

</div>

> [!NOTE]
>
> **AVISO**: Esse repositório está em constante evolução. Se você encontrar algum erro ou tiver sugestões, por favor, abra uma [issue](https://github.com/lpsm-dev/gtoc/issues/new/choose) ou envie um [pull request](https://github.com/lpsm-dev/gtoc/pulls).

<!-- START_TABLE_OF_CONTENTS -->

- [Visão Geral](#visão-geral)
  - [Objetivo](#objetivo)
  - [Contexto e Motivação](#contexto-e-motivação)
- [Implementação](#implementação)
  - [Pré-requisitos](#pré-requisitos)
  - [Instalação](#instalação)
  - [Uso](#uso)
- [Contribuição](#contribuição)
- [Versionamento](#versionamento)
- [Troubleshooting](#troubleshooting)
- [Show your support](#show-your-support)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- END_TABLE_OF_CONTENTS -->

# Visão Geral

O `gtoc` é uma CLI escrita em Go que gera e mantém atualizado o sumário (table of contents) de arquivos Markdown. Ele lê os headings do arquivo, monta um índice hierárquico com âncoras compatíveis com o GitHub e insere o resultado entre marcadores HTML, de forma idempotente: rodar o comando duas vezes produz o mesmo resultado.

## Objetivo

Eliminar a manutenção manual de sumários em READMEs e documentações longas. O `gtoc` cuida de:

- Gerar o índice a partir dos headings reais do arquivo (níveis `#` a `######`);
- Criar âncoras exatamente como o GitHub cria, incluindo acentos (`Instalação` vira `#instalação`) e headings duplicados (sufixos `-1`, `-2`);
- Ignorar headings dentro de blocos de código e do próprio sumário;
- Atualizar o bloco existente no lugar, preservando o restante do arquivo e as permissões.

## Contexto e Motivação

READMEs bem estruturados facilitam a navegação, mas sumários mantidos à mão ficam desatualizados a cada seção criada ou renomeada. Ferramentas existentes (como o doctoc) resolvem parte do problema, porém este projeto nasceu para: (1) ter uma solução em binário único, sem dependência de Node; (2) tratar corretamente âncoras com acentuação em português; e (3) servir de laboratório de boas práticas de desenvolvimento de CLIs em Go.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Implementação

## Pré-requisitos

Nenhum para usar o binário publicado nas releases. Para compilar a partir do código-fonte, é necessário Go `1.25+` (veja a versão exata no [go.mod](go.mod)).

## Instalação

Via `go install`:

```bash
go install github.com/lpsm-dev/gtoc@latest
```

Via binário: baixe o arquivo da sua plataforma na página de [releases](https://github.com/lpsm-dev/gtoc/releases) e coloque-o no seu `PATH`. Depois de instalado, atualize com o próprio CLI:

```bash
gtoc upgrade
```

Via Docker (build local):

```bash
docker build -t gtoc .
docker run --rm -v "$(pwd)":/work gtoc generate README.md
```

## Uso

Gerar ou atualizar o sumário de um arquivo:

```bash
gtoc generate README.md            # atualiza o arquivo no lugar
gtoc generate README.md --dry-run  # só mostra o que seria gerado
gtoc generate README.md --depth 3  # limita a profundidade dos headings
gtoc generate README.md --exclude "rascunho,privado"
```

O sumário é inserido (e depois atualizado) entre os marcadores abaixo. Na primeira execução sem marcadores, ele é adicionado no início do arquivo:

```markdown
<!-- START_TABLE_OF_CONTENTS -->
<!-- END_TABLE_OF_CONTENTS -->
```

Aplicar boas práticas de formatação ao README (marcadores `BEGIN_DOCS`/`END_DOCS`, âncora `readme-top` e links "back to top" ao fim de cada seção `#`):

```bash
gtoc analyze --file README.md
```

Flags do `generate`:

| Flag | Padrão | Descrição |
| --------- | ------ | ------------------------------------------------------------------ |
| `--file` | - | Caminho do arquivo Markdown (ou passe como argumento posicional) |
| `--depth` | `0` | Profundidade máxima de headings (`0` = ilimitado) |
| `--exclude` | - | Lista de textos de headings a excluir, separados por vírgula (match case-insensitive por substring) |
| `--dry-run` | `false` | Mostra o resultado sem escrever no arquivo |
| `--pretty` | `false` | No dry-run, renderiza o arquivo completo formatado no terminal |

Flags globais: `--log-level` (`debug`, `info`, `warn`, `error`, `fatal`), `--log-format` (`text`, `json`) e `--log-no-colors`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Contribuição

Gostaria de contribuir? Isso é ótimo! Temos um guia de contribuição para te ajudar. Clique [aqui](CONTRIBUTING.md) para lê-lo.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Versionamento

Para verificar o histórico de mudanças, acesse o arquivo [**CHANGELOG.md**](CHANGELOG.md).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Troubleshooting

Se você tiver algum problema, abra uma [issue](https://github.com/lpsm-dev/gtoc/issues/new/choose) nesse projeto.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Show your support

<div align="center">

Dê uma ⭐️ para este projeto se ele te ajudou!

<img src="https://github.com/lpsm-dev/lpsm-dev/blob/0062b174ec9877e6dfc78817f314b4a0690f63ff/.github/assets/yoda.gif" width="225"/>

<br>
<br>

Feito com 💜 pelo **Time de DevOps** :wave: inspirado no [readme-md-generator](https://github.com/kefranabg/readme-md-generator)

</div>

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- END_DOCS -->
