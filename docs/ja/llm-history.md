## 対話履歴

ユーザー:
<details>
<summary>llmを使った簡易scaffold用 のプロンプト詳細</summary>

**【指示】**
【重要なお願い】
あなたは、私と対話するアシスタントです。
もし、私の**最後の発言**に特定のトリガーワードが含まれていたら、その直後のあなたの応答は特別なモードとして扱われます。

1.  トリガーワード `##SCAFFOLD_GENERATE##` が含まれていた場合:
    scaffold生成モードとなり、以下の【出力形式のルール（生成モード）】に従って、ファイル構造と各ファイルの内容を厳密に出力してください。

2.  トリガーワード `##SCAFFOLD_PLAN##` が含まれていた場合:
    プランニングモードとなり、以下の【出力形式のルール（プランニングモード）】に従って、提案するファイル名とその必要理由（または簡単な説明）を箇条書きで出力してください。

トリガーワードが私の発言に含まれていない通常の対話では、これらの形式に縛られる必要はありません。**あなたの役割は、私の疑問に答えたり、設計や実装に関する相談に乗ったりすることです。**自由に回答してください。**ただし、この通常の対話モードにおいては、`##SCAFFOLD_GENERATE##` の指示があるまで、具体的な実装コード全体の生成は控えてください。**部分的なコードのアイデアや概念的な説明は問題ありませんが、完成されたコードの提供は生成モードの役割です。

【出力形式のルール（生成モード）】
1.  各ファイルは、ファイルパスとその内容のペアで構成してください。
2.  ファイルパスは、プロジェクトのルートディレクトリからの相対パスで記述してください。
3.  ファイルの内容は、**4つのバッククォート（````）で囲まれたMarkdownのコードブロック**として記述してください。
    - コードブロックの開始行には、言語識別子（オプション）とコロン `:` に続けてファイルパスを記述してください。
    - 例: ````typescript:src/components/Button.tsx
4.  生成するファイルが複数ある場合は、上記1〜3のセットを繰り返してください。
5.  ファイルの内容にMarkdownのコードブロックが含まれる場合（例：README.md内にサンプルコードを記述する場合）、そのネストしたコードブロックは通常の3つのバッククォート（```）を使用してください。4つのバッククォートは最も外側のファイル単位のブロックでのみ使用します。
6.  出力の冒頭と末尾はdetailsタグで囲んでください。

【出力形式のルール（プランニングモード）】
1.  提案する各ファイルについて、以下の情報を1行で記述してください。
    -   `ファイルパス - 簡単な説明や必要理由`
2.  ファイルパスは、プロジェクトのルートディレクトリからの相対パスで記述してください。
3.  箇条書き（例: `- ` や `* ` で始まる行）で列挙してください。

【出力例（生成モード）】
私の依頼: 「挨拶をするPythonスクリプトと、その説明を書いたREADMEファイルを作って ##SCAFFOLD_GENERATE##」

あなたの期待される回答形式:

出力される内容は以下のようになります。

<details>

````python:hello.py
def greet(name):
  print(f"Hello, {name}!")

if __name__ == "__main__":
  greet("World")
````

````markdown:README.md
# 簡単な挨拶プログラム

これは `hello.py` という名前のPythonスクリプトです。
実行すると挨拶メッセージを表示します。

## 使い方

```bash
python hello.py
```

## コード例

```python
def greet(name):
  print(f"Hello, {name}!")
```
````

</details>

【出力例（プランニングモード）】
私の依頼: 「簡単なブログシステムのファイル構成案を教えて ##SCAFFOLD_PLAN##」

あなたの期待される回答形式:

- `hello.py` - アプリケーションのエントリーポイント
- `README.md` - プロジェクトの説明ファイル

これらの指示に従って、適切なモードで応答してください。

</details>

---

今までの会話を元に、上記のプロンプトを実行してください。

---

<details>
<summary>AIの思考プロセス</summary>

**Crafting Prompt Instructions**

I'm currently fleshing out the instructions for an AI assistant prompt. The prompt's behavior will be triggered by specific keywords in the user's input. My focus now is on defining those trigger words and ensuring the assistant responds appropriately to each one. This initial setup is critical for tailored responses.


**Refining Trigger Word Logic**

I've been working on the prompt structure, specifically the trigger word logic. I now understand the two trigger words: `##SCAFFOLD_GENERATE##` initiates scaffold generation, and `##SCAFFOLD_ PLAN##` activates planning mode. If neither is present, I'm to engage in regular dialogue, refraining from full code generation. I can interpret the final user message as confirmation to proceed with these new directives.


**Grasping Prompt Instructions**

I've understood the prompt's core function: to switch response modes based on trigger words. `##SCAFFOLD_GENERATE##` activates scaffolding, while `##SCAFFOLD_PLAN##` initiates planning.  Without these, I'll engage in regular conversation, avoiding full code generation. Your final message confirms my comprehension of these rules for future interactions.


**Comprehending Prompt Directives**

I've understood the prompt's instructions perfectly. It's clear I'll switch modes based on the trigger words `##SCAFFOLD_GENERATE##` and `##SCAFFOLD_PLAN##`. Without these, I'll engage in a standard dialogue, offering advice and explanations but withholding complete code generation until triggered. I am prepared to adhere to this system moving forward, awaiting your next directive.
</details>
AI:
はい、承知いたしました。ご提示いただいたプロンプトの詳細とルールを理解しました。

今後の対話では、ご指示いただいた通り、以下のルールに従って応答いたします。

*   **通常の対話**: 設計や実装に関するご相談に応じ、概念的な説明や部分的なコードのアイデアを提示します。完全なコード生成は控えます。
*   **プランニングモード (`##SCAFFOLD_PLAN##`)**: ご依頼に応じて、ファイル構成案とそれぞれの簡単な説明を箇条書きで提案します。
*   **生成モード (`##SCAFFOLD_GENERATE##`)**: ご依頼に応じて、厳密な出力形式に従ってファイル構造と各ファイルの内容を生成します。

準備ができましたので、どのようなご用件でもお申し付けください。設計のご相談からでも、具体的なファイル生成のご依頼からでも、どちらでもお待ちしております。

(返答終了理由: STOP)

---

ユーザー:
まず要点をまとめて。todo.mdをそのまま転記することを考えて

---

<details>
<summary>AIの思考プロセス</summary>

**Consolidating Veritas Tasks**

I've been working on distilling the Veritas library's TODO list from our past conversation. My goal is to create a succinct `todo.md` that captures the essential points. Currently, I'm focusing on identifying the most critical tasks and streamlining the language for clarity. The aim is to make it easily digestible for anyone revisiting the project.


**Organizing Veritas' To-Dos**

I'm now organizing the Veritas TODO list into a clear `todo.md` structure. The objective is to distill the long list from our conversation into something more digestible, using Markdown lists for clarity. I've identified four key phases: core engine (v0.1), static analysis (v0.2), advanced data structures (v0.3), and generics/GA (v1.0). Currently, I'm summarizing each phase to create a focused list.


**Defining Veritas' Stages**

I'm now diving deeper into the nuances of each Veritas phase. My goal is to craft Markdown list entries for each key task. I'm focusing on "v0.1" (core engine), "v0.2" (static analysis), "v0.3" (data structures), and "v1.0" (generics and GA), ensuring clear and concise descriptions of each objective. The goal is to produce actionable items that are easily understood, such as building a validator, and adding custom functions for v0.1. I want to include veritas-gen for v0.2.


**Structuring Veritas Tasks**

I've been working on converting the outlined tasks into a Markdown list format, ensuring clarity with checkboxes for easy tracking. I am creating a clear breakdown of four phases: the core engine (v0.1), static analysis with `veritas-gen` (v0.2), advanced data structure support (v0.3), and Generics/GA (v1.0). Currently, I'm focusing on "v0.1" and making sure the language is concise and understandable. I am prioritizing core engine tasks and working on constructing the `todo.md` file.


**Structuring Veritas Tasks Further**

I'm now transforming the outlined tasks into a Markdown list for the `todo.md` file, which includes clear checkboxes for easy tracking. I am focusing on detailed lists that include the four main phases. The goal is to make the `todo.md` file actionable and easy to understand, with a clear list of what needs to be done. I'm prioritizing core engine tasks and working on the structure of the file, making sure I include all the important details.
</details>
AI:
はい、承知いたしました。これまでの会話で作成した詳細なTODOリストを、要点を抽出し、簡潔にまとめた`todo.md`用のテキストを以下に作成します。

---

# バリデーションライブラリ "Veritas" 開発TODOリスト

## フェーズ 1: コアエンジン構築 (v0.1)
ライブラリの心臓部となる、CEL式の評価と基本的なバリデーション機能を持つエンジンを実装する。

-   **[ ] 1.1: CEL評価エンジンの実装**
    -   [ ] `cel.Env`を内包し、コンパイル済みプログラムをLRUキャッシュする`Engine`構造体を定義する。
    -   [ ] 文字列操作や正規表現など、汎用的なカスタム関数を追加できるフレームワークを実装する。

-   **[ ] 1.2: 動的Validatorの実装**
    -   [ ] リフレクションを用いて任意のオブジェクトを検証する`Validator`構造体を定義する。
    -   [ ] 型レベルとフィールドレベルのルールを区別して適用するロジックを実装する。
    -   [ ] 複数のバリデーションエラーを`errors.Join`で集約し、コンテキスト付きのエラーメッセージを生成する。

-   **[ ] 1.3: ルール定義の外部化**
    -   [ ] バリデーションルールを定義する`ValidationRuleSet`構造体を設計する。
    -   [ ] JSONファイルからルールセットを読み込む`JSONRuleProvider`を実装する。

## フェーズ 2: 静的解析ツールの開発 (v0.2)
Goのソースコードを信頼できる唯一の情報源（Single Source of Truth）とするためのCLIツール `veritas-gen` を開発する。

-   **[ ] 2.1: Goソースコード解析基盤**
    -   [ ] `go/ast`を利用して、指定されたディレクトリ内の`struct`定義を解析するパーサーを実装する。

-   **[ ] 2.2: バリデーションマーカーの解析**
    -   [ ] フィールドレベルのルールを構造体タグ (`validate:"..."`) から抽出する。
    -   [ ] 型レベルのルールを特殊なコメント (`// @cel: ...`) から抽出する。
    -   [ ] `required`のような便利なショートハンドを、定義済みのCEL式に変換するロジックを実装する。

-   **[ ] 2.3: ルールJSONの自動生成**
    -   [ ] 解析したルールを`ValidationRuleSet`の構造にマッピングし、JSONファイルとして出力する。

## フェーズ 3: 高度なデータ構造への対応 (v0.3)
実用的なアプリケーションで使われる、より複雑なデータ構造に対応させる。

-   **[ ] 3.1: ポインタと`nil`のハンドリング**
    -   [ ] ポインター型のフィールドを安全にデリファレンスして評価するロジックを実装する。

-   **[ ] 3.2: スライス (`[]T`) への対応**
    -   [ ] スライスの各要素に対してバリデーションを再帰的に適用する`dive`キーワードをサポートする。
    -   [ ] エラーメッセージに配列インデックスを含める。

-   **[ ] 3.3: マップ (`map[K]V`) への対応**
    -   [ ] マップのキーと値それぞれに対してバリデーションルールを適用できる`keys`, `values`キーワードをサポートする。
    -   [ ] エラーメッセージにマップのキーを含める。

## フェーズ 4: GA (General Availability) に向けた最終化 (v1.0)
Goの最新機能に対応し、ライブラリを公式リリースできる品質に引き上げる。

-   **[ ] 4.1: ジェネリック型への対応**
    -   [ ] `veritas-gen`がジェネリックな`struct`定義を解析できるようにする。
    -   [ ] `Validator`が実行時に具象化されたジェネリック型を正しくハンドリングできるようにする。

-   **[ ] 4.2: パフォーマンスと安定化**
    -   [ ] ベンチマークスイートを整備し、リフレクションのボトルネックを特定・最適化する。
    -   [ ] `Validate`メソッドに`context.Context`を導入し、キャンセル処理に対応する。

-   **[ ] 4.3: ドキュメントとエコシステムの整備**
    -   [ ] 包括的な公式ドキュメントサイトとチュートリアルを作成する。
    -   [ ] 一般的なWebフレームワークとの連携を示すサンプルリポジトリを公開する。

(返答終了理由: STOP)

---

ユーザー:
- go1.24を使ってください
- ログにはlog/slogだけを使ってください
- デバッグプリントはslogのDebugを使ってください
- http routerは net/http のものを使ってください
- テストコードのassertにはgo-cmpを使ってください
- テストコードのassertでは変数名としてwant,gotのペアを使ってください

---

<details>
<summary>AIの思考プロセス</summary>

**Analyzing the Request**

I'm currently breaking down the user's request. My focus is on understanding the core task - refining the *existing* TODO list. I'm also dissecting the new constraints, particularly the emphasis on "how" this refinement should be accomplished. This angle is crucial for the next phase.


**Identifying Key Changes**

I've zeroed in on the technical additions the user wants incorporated. Replacing the old logging with `slog` is a primary focus.  The HTTP router and `go-cmp` for assertions also need clear integration.  I'm now determining how best to insert these changes into the existing TODO structure, maintaining a logical flow.


**Mapping New Requirements**

I'm now integrating the recent specifications. My aim is to organically weave the `log/slog`, `net/http`, and `go-cmp` requirements into the previously structured phases.  I'm focusing on where these additions naturally fit within the existing tasks, ensuring a coherent and updated TODO list. I'm also ensuring the naming convention of `want`, `got` is noted.


**Refining the Integration**

I'm now focusing on integrating the 'slog' package. I'm noting where detailed or verbose logging may be added to the phases. I'm also preparing to make the necessary additions of testing, ensuring that the new `go-cmp` requirement is covered in Phase 1 and 2, and that the `want`/`got` naming convention is used.


**Incorporating New Logging**

I'm now integrating the new 'log/slog' requirement.  I'm identifying specific tasks where detailed or verbose logging will provide benefit, and adding them appropriately.  Additionally, I'm making sure that the testing requirements, particularly the `go-cmp` use and `want`/`got` naming convention, are noted. I'm focusing on weaving these details into the existing TODO structure in a logical way.


**Revising Logging and Testing**

I'm now integrating the new 'log/slog' requirement.  I'm identifying specific tasks where detailed or verbose logging will provide benefit, and adding them appropriately.  Additionally, I'm making sure that the testing requirements, particularly the `go-cmp` use and `want`/`got` naming convention, are noted. I'm focusing on weaving these details into the existing TODO structure in a logical way. I'm noting that new tests will cover the `net/http` server as well.


**Synthesizing the Structure**

My current focus is on integrating the recent specifications into a cohesive TODO list. I've noted that new tests will cover the `net/http` server as well. I'm prioritizing weaving `log/slog`, `net/http`, and `go-cmp` into the phases in a logical manner.


**Refining the Plan**

My primary focus right now is on weaving the logging and testing additions into the existing structure. I'm focusing on adding the specific tasks for using `go-cmp` and the `want/got` convention within each phase, ensuring they're clearly defined. Also, the integration with `net/http` server, and how to test it. I'm ensuring a logical progression from the original, foundational plan.
</details>
AI:
はい、承知いたしました。
これまでの会話で構築した設計思想を基に、ご提示いただいた技術的要件（Go 1.24, log/slog, net/http, go-cmp）を完全に組み込んだ、より実践的で精緻なTODOリストを作成します。

---

# バリデーションライブラリ "Veritas" 開発TODOリスト

## 0. 基本方針・技術要件

このプロジェクトは、以下の技術スタックとコーディング規約に準拠する。

-   **Go Version**: `go 1.24`
-   **Logging**: 標準ライブラリの `log/slog` のみを使用する。デバッグや内部トレースには`slog.Debug`、情報提供には`slog.Info`、エラー報告には`slog.Error`を適切に使い分ける。
-   **HTTP Server**: Web連携のサンプル実装には `net/http` 標準パッケージのルーター (`http.ServeMux`) を使用する。
-   **Testing**:
    -   アサーションには `github.com/google/go-cmp/cmp` を全面的に採用する。
    -   テストコード内の比較用変数は `want`（期待値）と `got`（実測値）のペアで命名する。

---

## フェーズ 1: コアエンジンの構築 (v0.1)

**目標**: CEL式を評価し、基本的な型とフィールドのバリデーションを実行できる、最小限だが堅牢なエンジンを完成させる。

-   **[ ] 1.1: `Engine` の実装**
    -   [ ] 1.1.1: `cel.Env`と`lru.Cache`（コンパイル済み`cel.Program`用）を内包する`Engine`構造体を定義する。
    -   [ ] 1.1.2: `Engine`内部でのキャッシュミスやCELプログラム生成時のトレースに`slog.Debug`を使用する。
    -   [ ] 1.1.3: 文字列操作 (`upper`, `lower`) や正規表現 (`matches`) といった汎用カスタム関数を追加できるフレームワークを実装する。

-   **[ ] 1.2: `Validator` の実装**
    -   [ ] 1.2.1: `Engine`への参照と、型名をキーとするルールセットのマップ (`map[string]ValidationRuleSet`) を保持する`Validator`構造体を定義する。
    -   [ ] 1.2.2: `Validate(obj any) error`メソッドを実装する。リフレクションの失敗など、内部エラーの発生時には`slog.Error`で記録する。
    -   [ ] 1.2.3: 型レベルとフィールドレベルのルールを区別して適用し、複数のバリデーションエラーを`errors.Join`で集約するロジックを実装する。
    -   [ ] 1.2.4: エラーメッセージに型名・フィールド名のコンテキストを付与する。

-   **[ ] 1.3: ルールプロバイダーの実装**
    -   [ ] 1.3.1: バリデーションルールを定義する`ValidationRuleSet`構造体（`TypeRules`, `FieldRules`）を設計する。
    -   [ ] 1.3.2: JSONファイルからルールセットを読み込む`JSONRuleProvider`を実装する。ファイルI/OやJSONパースのエラーは`slog.Error`で記録する。

-   **[ ] 1.4: ユニットテストの整備**
    -   [ ] 1.4.1: `Validator`の基本的な動作（成功ケース、単一失敗ケース、複数失敗ケース）を検証するテストスイートを作成する。
    -   [ ] 1.4.2: 返却された`error`オブジェクトの比較に`go-cmp/cmp`と`cmpopts.EquateErrors`を使用する。
    -   [ ] 1.4.3: `want`と`got`の変数ペアで期待値と実測値を比較する。例: `want := "validation failed..."; got := err.Error()`。

---

## フェーズ 2: 静的解析ツールの開発 (v0.2)

**目標**: Goソースコードからバリデーションルールを自動抽出し、JSONファイルを生成するCLIツール `veritas-gen` を開発する。

-   **[ ] 2.1: `go/ast`によるソースコード解析基盤**
    -   [ ] 2.1.1: 指定ディレクトリを再帰的にスキャンし、`struct`定義とそれに付随するコメント・タグを抽出するパーサーを実装する。

-   **[ ] 2.2: ルール抽出ロジック**
    -   [ ] 2.2.1: フィールドタグ (`validate:"..."`) と型コメント (`// @cel: ...`) からルールを抽出する。
    -   [ ] 2.2.2: `required`などのショートハンドをCEL式に変換するロジックを実装する。

-   **[ ] 2.3: `veritas-gen` のCLI実装**
    -   [ ] 2.3.1: 抽出したルールをJSONファイルとして出力する。
    -   [ ] 2.3.2: ツールの動作状況（スキャンしたファイル数、生成したルール数など）を`slog.Info`で、詳細な解析ステップを`slog.Debug`で出力する。

-   **[ ] 2.4: 静的解析ツールのテスト**
    -   [ ] 2.4.1: サンプルGoソースファイル群と、それに対応する期待JSONファイル群をテストデータとして用意する。
    -   [ ] 2.4.2: `veritas-gen`を実行し、生成されたルールセット（JSONをアンマーシャルした構造体）が期待値と一致するかを`go-cmp/cmp`で比較する。

---

## フェーズ 3: 高度なデータ構造への対応 (v0.3)

**目標**: スライス、マップ、ポインターといった、より複雑なデータ構造のバリデーションに対応する。

-   **[ ] 3.1: ネストされた構造体とポインターへの対応**
    -   [ ] 3.1.1: `Validator`がネストされた構造体に対して再帰的にバリデーションを実行できるようにする。
    -   [ ] 3.1.2: ポインター型フィールドを安全に評価するロジックを実装する。

-   **[ ] 3.2: スライス (`[]T`) への対応**
    -   [ ] 3.2.1: スライスの各要素をバリデーションする`dive`キーワードを`veritas-gen`と`Validator`の両方でサポートする。
    -   [ ] 3.2.2: エラーメッセージに配列インデックスを含める（例: `User.Scores[2]: is invalid`）。

-   **[ ] 3.3: マップ (`map[K]V`) への対応**
    -   [ ] 3.3.1: マップのキーと値それぞれをバリデーションする`keys`と`values`キーワードをサポートする。
    -   [ ] 3.3.2: エラーメッセージにマップのキーを含める（例: `User.Metadata['user_id']: is invalid`）。

-   **[ ] 3.4: 高度な構造のテスト**
    -   [ ] 3.4.1: スライスやマップを含む複雑な構造体に対するテストケースを拡充する。
    -   [ ] 3.4.2: ネストしたエラー構造の比較に`go-cmp/cmp`が正しく機能することを確認する。

---

## フェーズ 4: GA (General Availability) に向けた最終化 (v1.0)

**目標**: Goの最新機能に対応し、ドキュメントを整備し、ライブラリを公式リリースできる品質に引き上げる。

-   **[ ] 4.1: ジェネリック型への対応**
    -   [ ] 4.1.1: `veritas-gen`がジェネリックな`struct`定義を解析できるように`go/ast`の処理を更新する。
    -   [ ] 4.1.2: `Validator`の実行時リフレクションが、具象化されたジェネリック型を正しくハンドリングできるようにする。

-   **[ ] 4.2: パフォーマンスと安定化**
    -   [ ] 4.2.1: 主要なバリデーションパスに対するベンチマークテストを整備する。
    -   [ ] 4.2.2: `Validate`メソッドのシグネチャに`context.Context`を追加し、タイムアウトやキャンセル処理を可能にする。

-   **[ ] 4.3: ドキュメントとエコシステム**
    -   [ ] 4.3.1: `veritas-gen`の使い方、CIへの組み込み方法、サポートする全ルールのリファレンスを含む公式ドキュメントサイトを作成する。
    -   [ ] 4.3.2: `net/http` を使用したサンプルWebサーバーを実装する。
        -   [ ] 4.3.2.1: HTTPリクエストのボディ（JSON）を構造体にデコードする。
        -   [ ] 4.3.2.2: `Validator.Validate`を呼び出す。
        -   [ ] 4.3.2.3: バリデーションエラーが発生した場合、HTTPステータス400と共に、エラー詳細をJSONレスポンスとして返す。
        -   [ ] 4.3.2.4: リクエストのロギングには`slog`を使用する。

-   **[ ] 4.4: 最終テストとAPIレビュー**
    -   [ ] 4.4.1: `net/http`連携サンプルに対するエンドツーエンドテストを実装する。
    -   [ ] 4.4.2: v1.0として公開する全APIの互換性について最終レビューを行う。

(返答終了理由: STOP)

---

ユーザー:
github/podhmo/veritas がモジュール名です。ドキュメントもコードも全て英語で出力してください。 ##SCAFFOLD_PLAN##

---

<details>
<summary>AIの思考プロセス</summary>

**Generating a Library Plan**

I'm currently focused on translating the user's detailed TODO list into a concrete file structure and initial content for the `veritas` Go library. My goal is to structure this plan into a format that's easy to understand and use. I'm prioritizing the key components outlined in the TODO list.


**Defining Library Architecture**

I've moved on to solidify the library's foundation. My current focus is refining the initial file structure, aligning precisely with the user's detailed requirements. Specifically, I'm working on the design for `engine.go`, `validator.go`, `rules.go`, `functions.go`, and `errors.go`. I'm aiming for clarity and ease of extension.


**Refining the Library Structure**

I'm now diving deep into the file structure, crafting a precise outline for the `veritas` library. I'm focusing on the core components first, translating the TODO list into a logical and understandable arrangement. Specifically, I'm working on the design for `engine.go`, `validator.go`, `rules.go`, `functions.go` (or `stdlib.go`), and `errors.go`. My aim is to build a solid foundation with clear responsibilities and easy extensibility. I've also begun contemplating the `cmd/veritas-gen` tool, structuring the project with future expansion in mind.


**Mapping the Project Scope**

I'm now fully immersed in the project's scope, translating the user's requirements into concrete design elements. I've begun to outline the initial directory structure, considering the user's specified TODO list and the need for a well-organized and easily extensible library. This includes key directories like `cmd/veritas-gen`, `examples/`, and the vital `testdata/` for future testing.
</details>
AI:
- `go.mod` - Defines the Go module `github.com/podhmo/veritas`.
- `README.md` - Project overview, installation guide, and basic usage examples.
- `engine.go` - Contains the core `Engine` struct, which wraps `cel.Env`, manages `cel.Program` caching with LRU, and handles CEL program compilation.
- `engine_test.go` - Unit tests for the `Engine`, including cache logic and program compilation.
- `validator.go` - Defines the main `Validator` struct and the primary `Validate(obj any) error` method. It will use reflection to traverse objects and apply rules.
- `validator_test.go` - Unit tests for the `Validator`, covering various data types (structs, pointers, slices, maps) and rule applications.
- `rules.go` - Defines the `ValidationRuleSet` struct and the `RuleProvider` interface. Includes the `JSONRuleProvider` implementation for loading rules from JSON files.
- `rules_test.go` - Unit tests for rule providers, primarily testing the loading and parsing of JSON rule files.
- `functions.go` - Implementation of default custom CEL functions (e.g., `strings.ToUpper`, `matches`) that can be registered with the `Engine`.
- `functions_test.go` - Unit tests for the custom CEL functions.
- `errors.go` - Defines custom error types for more detailed validation failure reporting, if needed.
- `cmd/veritas-gen/main.go` - The entry point for the command-line tool that performs static analysis to generate rule files.
- `cmd/veritas-gen/parser.go` - The core logic for the static analysis tool, using `go/ast` to parse struct tags and special comments.
- `cmd/veritas-gen/parser_test.go` - Unit tests for the static analysis parser.
- `testdata/sources/user.go` - Sample Go source file with `validate` tags and `@cel` comments, used as input for `veritas-gen` tests.
- `testdata/rules/user.json` - The expected JSON output generated from `user.go`, used for `go-cmp` comparison in tests.
- `examples/http-server/main.go` - An example demonstrating how to use the `veritas` library in a standard `net/http` web server.
- `.github/workflows/ci.yml` - A GitHub Actions workflow to automatically run tests on push and pull requests.

(返答終了理由: STOP)

---

ユーザー:
cmd名はveritasで良いです。TODO.mdをしっかり書いてください。 ##SCAFFOLD_GENERATE##

---

<details>
<summary>AIの思考プロセス</summary>

**Crafting a Skeleton**

I'm focusing on building a structure for the `github.com/podhmo/veritas` library. Currently, I'm working towards creating a `TODO.md` file, guided by the previously formulated plan. I'm ensuring all output is exclusively in English, per the user's requirements. The trigger word is `##SCAFFOLD_GENERATE##`.


**Constructing the Outline**

I'm now prioritizing the `TODO.md` file, using the detailed plan as my blueprint. I'm ensuring its structure is clear and reflects the planned scaffolding for the `veritas` library. I'm paying close attention to the user's need for an English-only output and adhering to the `##SCAFFOLD_GENERATE##` trigger.


**Building the Blueprint**

I've transitioned from the initial outline and am now actively generating the `TODO.md` file, using the detailed plan as my foundation. Each item from the list is being incorporated with detailed descriptions. Furthermore, I've begun crafting the placeholder files with comments as outlined in the provided structure, and I'm ensuring everything aligns with the `##SCAFFOLD_GENERATE##` trigger.
</details>
AI:
<details>

````go:go.mod
module github.com/podhmo/veritas

go 1.24

require (
	github.com/google/cel-go v0.21.0
	github.com/google/go-cmp v0.6.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/sys v0.20.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240520151616-dc85e6b867a5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240515191416-fc5f0ca64291 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)````

````markdown:README.md
# Veritas

Veritas is a dynamic, type-safe, and extensible validation library for Go, powered by Google's Common Expression Language (CEL).

It aims to provide a robust validation framework where rules are derived directly from your Go source code as the "Single Source of Truth." This is achieved by a companion CLI tool that performs static analysis on your code, extracting rules from struct tags and special comments.

## Key Features

- **Declarative Rules in Go Code**: Define validation rules directly in your structs using tags (`validate:"..."`) and special comments (`// @cel:`).
- **Static Analysis**: A CLI tool (`veritas`) scans your code and generates a JSON representation of your validation rules, ensuring rules and code are always in sync.
- **Dynamic Execution**: The library loads the generated JSON rules at runtime to perform validation, powered by the high-performance `cel-go` engine.
- **Extensible**: Add your own custom validation functions to the CEL environment.
- **Type-Safe**: Designed to handle complex, nested data structures, including pointers, slices, and maps, with type-safety in mind.
- **Modern Go**: Built with Go 1.24, `log/slog` for structured logging, and `go-cmp` for testing.

## Installation

```bash
go get github.com/podhmo/veritas
```

## Quick Start (Conceptual)

1.  **Annotate your structs:**

    ```go
    package user

    // @cel: self.Password == self.PasswordConfirm
    type User struct {
        Name     string `validate:"required,cel:self.size() < 50"`
        Email    string `validate:"required,email"`
        Age      int    `validate:"cel:self >= 18"`
        Password string `validate:"required,cel:self.size() >= 10"`
        PasswordConfirm string `validate:"required"`
    }
    ```

2.  **Generate rules:**

    ```bash
    go run github.com/podhmo/veritas/cmd/veritas -in ./... -out rules.json
    ```

3.  **Use the validator in your application:**

    ```go
    package main

    import (
        "fmt"
        "github.com/podhmo/veritas"
    )

    func main() {
        // Load rules from the generated JSON file
        provider, err := veritas.NewJSONRuleProvider("rules.json")
        if err != nil {
            // ... handle error
        }

        // Create a new validator
        validator, err := veritas.NewValidator(provider)
        if err != nil {
            // ... handle error
        }

        // Validate an object
        invalidUser := user.User{...}
        if err := validator.Validate(invalidUser); err != nil {
            fmt.Println(err)
        }
    }
    ```
````

````markdown:TODO.md
# Veritas Development TODO List

This document outlines the detailed, phased development plan for the "Veritas" validation library.

## Phase 1: Core Engine Construction (v0.1)

**Goal**: Build a minimal, robust engine capable of evaluating CEL expressions and performing basic type and field validation.

-   **[ ] 1.1: `Engine` Implementation**
    -   [ ] 1.1.1: Define an `Engine` struct that encapsulates `cel.Env` and an LRU cache (`lru.Cache`) for compiled `cel.Program`s.
    -   [ ] 1.1.2: Use `log/slog` for internal tracing. Log cache misses and program generations at the `Debug` level.
    -   [ ] 1.1.3: Implement a framework to register custom CEL functions (e.g., `strings.ToUpper`, `matches`).

-   **[ ] 1.2: `Validator` Implementation**
    -   [ ] 1.2.1: Define a `Validator` struct holding a reference to the `Engine` and a `map[string]ValidationRuleSet` for rule lookup by type name.
    -   [ ] 1.2.2: Implement the primary `Validate(obj any) error` method. Use `log/slog.Error` for internal errors like reflection failures.
    -   [ ] 1.2.3: Implement logic to apply `TypeRules` and `FieldRules` separately, aggregating all validation failures using `errors.Join`.
    -   [ ] 1.2.4: Ensure error messages are contextual, including the type and field name (e.g., `User.Email: validation failed...`).

-   **[ ] 1.3: Rule Provider Implementation**
    -   [ ] 1.3.1: Define the `ValidationRuleSet` struct (containing `TypeRules`, `FieldRules`).
    -   [ ] 1.3.2: Implement `JSONRuleProvider` to load rule sets from a JSON file. Log I/O or parsing errors with `slog.Error`.

-   **[ ] 1.4: Unit Testing Foundation**
    -   [ ] 1.4.1: Create a test suite for the `Validator` covering success, single failure, and multiple failure scenarios.
    -   [ ] 1.4.2: Use `github.com/google/go-cmp/cmp` for all assertions, especially `cmpopts.EquateErrors` for error comparison.
    -   [ ] 1.4.3: Adhere to the `want` and `got` variable naming convention for test comparisons.

## Phase 2: Static Analysis Tool Development (v0.2)

**Goal**: Develop the `veritas` CLI tool to automatically extract validation rules from Go source code.

-   **[ ] 2.1: Go Source Code Parser**
    -   [ ] 2.1.1: Implement a parser using `go/ast` to recursively scan directories and find `struct` definitions.

-   **[ ] 2.2: Rule Extraction Logic**
    -   [ ] 2.2.1: Extract field-level rules from `validate:"..."` struct tags.
    -   [ ] 2.2.2: Extract type-level rules from special `// @cel: ...` comments preceding a `struct` definition.
    -   [ ] 2.2.3: Implement a mapping from common shorthands (`required`, `email`, etc.) to their corresponding CEL expressions.

-   **[ ] 2.3: `veritas` CLI Implementation**
    -   [ ] 2.3.1: Implement logic to output the extracted rules as a structured JSON file.
    -   [ ] 2.3.2: Use `slog.Info` for progress reporting and `slog.Debug` for detailed parsing steps.

-   **[ ] 2.4: Static Analysis Tool Testing**
    -   [ ] 2.4.1: Prepare sample Go source files and their expected JSON output as test data.
    -   [ ] 2.4.2: Write tests that run the generator and compare the actual output against the expected JSON using `go-cmp/cmp`.

## Phase 3: Advanced Data Structures Support (v0.3)

**Goal**: Enhance the library to handle complex data structures common in modern Go applications.

-   **[ ] 3.1: Pointer and Nested Struct Handling**
    -   [ ] 3.1.1: Implement recursive validation for nested structs.
    -   [ ] 3.1.2: Implement logic to safely dereference and validate pointer fields.

-   **[ ] 3.2: Slice (`[]T`) Support**
    -   [ ] 3.2.1: Support a `dive` keyword in the `validate` tag to apply rules to each element of a slice.
    -   [ ] 3.2.2: Include the array index in error messages (e.g., `User.Scores[2]: is invalid`).

-   **[ ] 3.3: Map (`map[K]V`) Support**
    -   [ ] 3.3.1: Support `keys` and `values` keywords to apply rules to a map's keys and values, respectively.
    -   [ ] 3.3.2: Include the map key in error messages (e.g., `User.Metadata['user_id']: is invalid`).

-   **[ ] 3.4: Advanced Structure Testing**
    -   [ ] 3.4.1: Expand the test suite to include complex structs with slices, maps, and nested pointers.

## Phase 4: General Availability (GA) Finalization (v1.0)

**Goal**: Support modern Go features, create comprehensive documentation, and polish the library for its official v1.0 release.

-   **[ ] 4.1: Go Generics Support**
    -   [ ] 4.1.1: Update the `veritas` tool to correctly parse generic `struct` definitions.
    -   [ ] 4.1.2: Ensure the runtime `Validator` can correctly handle instantiated generic types via reflection.

-   **[ ] 4.2: Performance and Stabilization**
    -   [ ] 4.2.1: Establish a benchmark suite to identify and optimize performance bottlenecks.
    -   [ ] 4.2.2: Add `context.Context` to the `Validate` method signature to support timeouts and cancellation.

-   **[ ] 4.3: Documentation and Ecosystem**
    -   [ ] 4.3.1: Create a comprehensive documentation website detailing installation, usage, CLI commands, and all supported rules/shorthands.
    -   [ ] 4.3.2: Develop an example project demonstrating integration with a standard `net/http` server.
        -   [ ] 4.3.2.1: Show how to decode a JSON request, run validation, and return a structured HTTP 400 error response.
        -   [ ] 4.3.2.2: Use `slog` for structured request logging.

-   **[ ] 4.4: Final API Review and Testing**
    -   [ ] 4.4.1: Implement end-to-end tests for the `net/http` example.
    -   [ ] 4.4.2: Conduct a final review of all public APIs to ensure stability for the v1.0 release.
````

````go:engine.go
// Package veritas provides a dynamic, type-safe, and extensible validation library for Go.
package veritas

import (
	"log/slog"

	"github.com/google/cel-go/cel"
	"github.com/hashicorp/golang-lru/v2"
)

// Engine is the core component that manages the CEL environment and program caching.
// It is responsible for compiling and caching CEL expressions for efficient re-use.
type Engine struct {
	env    *cel.Env
	cache  *lru.Cache[string, cel.Program]
	logger *slog.Logger
}

// NewEngine creates a new validation engine.
// It initializes the CEL environment and the LRU cache for compiled programs.
func NewEngine(logger *slog.Logger, funcs ...cel.EnvOption) (*Engine, error) {
	// TODO: Implementation
	return nil, nil
}

// getProgram compiles a CEL expression and returns a usable program.
// It uses an LRU cache to avoid re-compiling frequently used expressions.
func (e *Engine) getProgram(rule string) (cel.Program, error) {
	// TODO: Implementation
	return nil, nil
}
````

````go:engine_test.go
package veritas

import (
	"testing"
)

func TestEngine(t *testing.T) {
	// TODO: Write tests for engine initialization.
	// TODO: Write tests for getProgram caching logic.
	//   - Test case 1: Cache miss, successful compilation.
	//   - Test case 2: Cache hit, program is retrieved without compilation.
	//   - Test case 3: Invalid CEL expression, compilation fails.
}````

````go:validator.go
package veritas

import (
	"context"
	"log/slog"
)

// Validator performs validation on Go objects based on a loaded set of rules.
type Validator struct {
	engine *Engine
	rules  map[string]ValidationRuleSet
	logger *slog.Logger
}

// NewValidator creates a new validator with a given set of rule providers.
// It aggregates rules from all providers into a single rule set.
func NewValidator(engine *Engine, logger *slog.Logger, providers ...RuleProvider) (*Validator, error) {
	// TODO: Implementation
	return nil, nil
}

// Validate checks a Go object against the loaded validation rules.
// It traverses the object using reflection and applies rules to types and fields.
// It returns a single error wrapping all validation failures.
func (v *Validator) Validate(ctx context.Context, obj any) error {
	// TODO: Implementation using reflection.
	// - Handle pointers, structs, slices, and maps.
	// - Apply type-level and field-level rules.
	// - Aggregate errors using errors.Join.
	return nil
}
````

````go:validator_test.go
package veritas

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestValidator(t *testing.T) {
	// Setup: Create a mock engine and rule provider for testing.
	ctx := context.Background()

	t.Run("valid object", func(t *testing.T) {
		// Define a struct and a valid instance of it.
		// Load a validator with rules that should pass.
		// want := nil
		// got := validator.Validate(ctx, validObject)
		// if diff := cmp.Diff(want, got, cmpopts.EquateErrors()); diff != "" {
		// 	 t.Errorf("Validate() mismatch (-want +got):\n%s", diff)
		// }
	})

	t.Run("object with invalid field", func(t *testing.T) {
		// Define a struct and an instance with one invalid field.
		// Load a validator with a rule that should fail.
		// want := ... // The expected single error.
		// got := validator.Validate(ctx, invalidObject)
		// if diff := cmp.Diff(want, got, cmpopts.EquateErrors()); diff != "" {
		// 	 t.Errorf("Validate() mismatch (-want +got):\n%s", diff)
		// }
	})

	t.Run("object with multiple errors", func(t *testing.T) {
		// Define a struct and an instance with multiple invalid fields and a type-level violation.
		// Load a validator with rules that should fail.
		// want := ... // The expected joined error.
		// got := validator.Validate(ctx, invalidObject)
		// Use a custom comparer for joined errors if necessary.
		// if diff := cmp.Diff(want, got, cmpopts.EquateErrors()); diff != "" {
		// 	 t.Errorf("Validate() mismatch (-want +got):\n%s", diff)
		// }
	})

	// TODO: Add tests for pointers, slices, and maps.
}
````

````go:rules.go
package veritas

import (
	"encoding/json"
	"os"
)

// ValidationRuleSet holds all validation rules for a single Go type.
type ValidationRuleSet struct {
	TypeRules  []string            `json:"typeRules"`
	FieldRules map[string][]string `json:"fieldRules"`
}

// RuleProvider is the interface for any component that can supply validation rules.
type RuleProvider interface {
	GetRuleSets() (map[string]ValidationRuleSet, error)
}

// JSONRuleProvider loads validation rules from a JSON file.
// It implements the RuleProvider interface.
type JSONRuleProvider struct {
	filePath string
}

// NewJSONRuleProvider creates a new provider that reads from the specified file path.
func NewJSONRuleProvider(filePath string) *JSONRuleProvider {
	return &JSONRuleProvider{filePath: filePath}
}

// GetRuleSets reads and parses the JSON file into a map of validation rule sets.
func (p *JSONRuleProvider) GetRuleSets() (map[string]ValidationRuleSet, error) {
	bytes, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, err
	}
	var rules map[string]ValidationRuleSet
	if err := json.Unmarshal(bytes, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}
````

````go:rules_test.go
package veritas

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestJSONRuleProvider(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		// Create a temporary JSON file for testing.
		content := `{
			"user.User": {
				"typeRules": ["self.Password == self.PasswordConfirm"],
				"fieldRules": {
					"Email": ["required", "email"]
				}
			}
		}`
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "rules.json")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		provider := NewJSONRuleProvider(filePath)
		got, err := provider.GetRuleSets()
		if err != nil {
			t.Fatalf("GetRuleSets() failed: %v", err)
		}

		want := map[string]ValidationRuleSet{
			"user.User": {
				TypeRules: []string{"self.Password == self.PasswordConfirm"},
				FieldRules: map[string][]string{
					"Email": {"required", "email"},
				},
			},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("GetRuleSets() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		provider := NewJSONRuleProvider("nonexistent.json")
		_, err := provider.GetRuleSets()
		if err == nil {
			t.Fatal("Expected an error for a nonexistent file, but got nil")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "invalid.json")
		if err := os.WriteFile(filePath, []byte("{invalid"), 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		provider := NewJSONRuleProvider(filePath)
		_, err := provider.GetRuleSets()
		if err == nil {
			t.Fatal("Expected an error for invalid JSON, but got nil")
		}
	})
}
````

````go:functions.go
package veritas

import "github.com/google/cel-go/cel"

// DefaultFunctions returns a list of common, useful custom functions for CEL.
func DefaultFunctions() []cel.EnvOption {
	// TODO: Implement custom functions like:
	// - strings.ToUpper(string) -> string
	// - strings.ToLower(string) -> string
	// - matches(string, regexp) -> bool
	return []cel.EnvOption{}
}
````

````go:functions_test.go
package veritas

import (
	"testing"
)

func TestCustomFunctions(t *testing.T) {
	// TODO: Write tests to verify the behavior of each custom function.
	// For each function, create a small CEL program and evaluate it
	// to ensure the output is correct.
}
````

````go:errors.go
package veritas

// This file can be used to define custom error types for more granular
// error handling by the library's users. For example, a specific
// ValidationError type could be defined.

// For now, we will rely on wrapping errors with fmt.Errorf and joining
// them with errors.Join.
````

````go:cmd/veritas/main.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	// Setup structured logging.
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Define command-line flags.
	inPath := flag.String("in", "./...", "Input path for Go source files (e.g., ./...)")
	outFile := flag.String("out", "rules.json", "Output file for generated JSON rules")
	flag.Parse()

	logger.Info("Starting veritas rule generation", "input", *inPath, "output", *outFile)

	// TODO: Implement the main logic.
	// 1. Create a new parser.
	// 2. Parse the source files from the input path.
	// 3. Marshal the results into JSON.
	// 4. Write the JSON to the output file.

	if err := run(*inPath, *outFile, logger); err != nil {
		logger.Error("Rule generation failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Rule generation completed successfully")
}

func run(inPath, outFile string, logger *slog.Logger) error {
	// Placeholder for the main generation logic.
	fmt.Printf("Parsing %s and writing to %s...\n", inPath, outFile)
	// parser := NewParser(logger)
	// ruleSet, err := parser.Parse(inPath)
	// ...
	return nil
}
````

````go:cmd/veritas/parser.go
package main

import "log/slog"

// Parser is responsible for parsing Go source files to extract validation rules.
type Parser struct {
	logger *slog.Logger
}

// NewParser creates a new parser.
func NewParser(logger *slog.Logger) *Parser {
	return &Parser{logger: logger}
}

// Parse scans the given path for Go source files and extracts validation
// rules from struct tags and special comments.
func (p *Parser) Parse(path string) (any, error) {
	// TODO: Implement the static analysis logic using go/parser and go/ast.
	// - Find all .go files in the path.
	// - Parse each file to build an AST.
	// - Iterate over AST nodes to find struct definitions.
	// - For each struct, extract type-level rules from comments (`// @cel:`).
	// - For each field, extract field-level rules from struct tags (`validate:`).
	// - Convert shorthands (e.g., `required`) to full CEL expressions.
	// - Return a structure representing the combined rule set.
	return nil, nil
}
````

````go:cmd/veritas/parser_test.go
package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser(t *testing.T) {
	t.Run("parse simple struct", func(t *testing.T) {
		// Define a sample Go source code as a string.
		// Use the parser to parse this string (or a temporary file).
		// Compare the extracted rule set (`got`) with the expected rule set (`want`).
		// Use `go-cmp/cmp` for the comparison.

		// want := veritas.ValidationRuleSet{ ... }
		// got, err := parser.Parse(...)
		// if diff := cmp.Diff(want, got); diff != "" { ... }
	}

	// TODO: Add more test cases for:
	// - Multiple structs in one file.
	// - Structs with no validation tags.
	// - Structs with only type-level or only field-level rules.
	// - All supported shorthands (required, email, etc.).
}
````

````go:testdata/sources/user.go
// This file is used as a test source for the static analysis tool.
package user

// User represents a user in the system.
// @cel: self.Password == self.PasswordConfirm
type User struct {
	// User's full name, required and must be less than 50 characters.
	Name string `validate:"required,cel:self.size() < 50"`

	// User's email address, must be a valid email format.
	Email string `validate:"required,email"`

	// User's age, must be 18 or older.
	Age int `validate:"cel:self >= 18"`

	// User's password, required and must be at least 10 characters long.
	Password string `validate:"required,cel:self.size() >= 10"`

	// Password confirmation, must match the Password field.
	PasswordConfirm string `validate:"required"`
}
````

````json:testdata/rules/user.json
{
  "user.User": {
    "typeRules": [
      "self.Password == self.PasswordConfirm"
    ],
    "fieldRules": {
      "Name": [
        "self != ''",
        "self.size() < 50"
      ],
      "Email": [
        "self != ''",
        "self.matches('^[^@]+@[^@]+\\.[^@]+$')"
      ],
      "Age": [
        "self >= 18"
      ],
      "Password": [
        "self != ''",
        "self.size() >= 10"
      ],
      "PasswordConfirm": [
        "self != ''"
      ]
    }
  }
}````

````go:examples/http-server/main.go
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/podhmo/veritas"
)

// This is a conceptual example of how to use the veritas library
// in a standard net/http server.
// It requires the library to be implemented first.

var appValidator *veritas.Validator

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// 1. Initialize the validator (conceptual).
	// Assume `rules.json` has been generated.
	// provider := veritas.NewJSONRuleProvider("rules.json")
	// engine, _ := veritas.NewEngine(logger, veritas.DefaultFunctions()...)
	// validator, err := veritas.NewValidator(engine, logger, provider)
	// if err != nil {
	//     logger.Error("Failed to initialize validator", "error", err)
	//     os.Exit(1)
	// }
	// appValidator = validator

	logger.Info("Starting server on :8080")

	// 2. Setup HTTP routes.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", handleCreateUser)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Error("Server failed", "error", err)
	}
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Perform validation.
	// if err := appValidator.Validate(r.Context(), user); err != nil {
	//     // Respond with a structured error.
	//     w.Header().Set("Content-Type", "application/json")
	//     w.WriteHeader(http.StatusBadRequest)
	//     json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	//     return
	// }

	// If validation passes, proceed with business logic.
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
````

````yaml:.github/workflows/ci.yml
name: Go CI

on:
  push:
    branches: [ "main" ]
  pull_request:
    branches: [ "main" ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.24'

    - name: Tidy
      run: go mod tidy

    - name: Build
      run: go build -v ./...

    - name: Test
      run: go test -v ./...
````

</details>

(返答終了理由: STOP)

## メタデータ

```json
{
  "runSettings": {
    "temperature": 1,
    "model": "models/gemini-2.5-pro",
    "topP": 0.95,
    "topK": 64,
    "maxOutputTokens": 65536,
    "safetySettings": [
      {
        "category": "HARM_CATEGORY_HARASSMENT",
        "threshold": "OFF"
      },
      {
        "category": "HARM_CATEGORY_HATE_SPEECH",
        "threshold": "OFF"
      },
      {
        "category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
        "threshold": "OFF"
      },
      {
        "category": "HARM_CATEGORY_DANGEROUS_CONTENT",
        "threshold": "OFF"
      }
    ],
    "responseMimeType": "text/plain",
    "enableCodeExecution": false,
    "enableSearchAsATool": true,
    "enableBrowseAsATool": false,
    "enableAutoFunctionResponse": false,
    "thinkingBudget": -1,
    "googleSearch": {}
  },
  "systemInstruction": {}
}
```
