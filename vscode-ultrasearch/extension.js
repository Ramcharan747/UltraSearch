const vscode = require('vscode');
const { exec } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

function findBinary(workspaceFolder) {
    const candidates = [
        path.join(workspaceFolder, 'ultrasearch'),
        path.join(workspaceFolder, '..', 'ultrasearch'),
        path.join(__dirname, 'ultrasearch'),
        path.join(__dirname, '..', 'ultrasearch')
    ];
    for (const c of candidates) {
        if (fs.existsSync(c)) {
            return c;
        }
    }
    return 'ultrasearch';
}

function activate(context) {
    let disposable = vscode.commands.registerCommand('ultrasearch.search', async function () {
        const query = await vscode.window.showInputBox({
            prompt: "Enter search query for UltraSearch (Deep Web Search)",
            placeHolder: "e.g., python playwright stealth bypass"
        });

        if (!query) return;

        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "UltraSearch: Extracting web data...",
            cancellable: false
        }, async (progress) => {
            return new Promise((resolve, reject) => {
                const tempFile = path.join(os.tmpdir(), `ultra_${Date.now()}.txt`);
                
                const workspaceFolder = vscode.workspace.workspaceFolders ? vscode.workspace.workspaceFolders[0].uri.fsPath : __dirname;
                const binaryPath = findBinary(workspaceFolder);
                
                const cmd = `"${binaryPath}" -query "${query}" -no-ai -output-format=llm-dense -output="${tempFile}"`;
                
                exec(cmd, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`UltraSearch Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    if (fs.existsSync(tempFile)) {
                        vscode.workspace.openTextDocument(tempFile).then(doc => {
                            vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                        });
                    } else {
                        vscode.window.showInformationMessage("Search completed but no output file was generated.");
                    }
                    resolve();
                });
            });
        });
    });

    let fastDisposable = vscode.commands.registerCommand('ultrasearch.fastSearch', async function () {
        const query = await vscode.window.showInputBox({
            prompt: "Enter search query for AI Overview & 10 URLs",
            placeHolder: "e.g., speed of light in vacuum"
        });

        if (!query) return;

        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "UltraSearch: Fetching AI Overview & URLs...",
            cancellable: false
        }, async (progress) => {
            return new Promise((resolve, reject) => {
                const tempFile = path.join(os.tmpdir(), `ultra_fast_${Date.now()}.txt`);
                const workspaceFolder = vscode.workspace.workspaceFolders ? vscode.workspace.workspaceFolders[0].uri.fsPath : __dirname;
                const binaryPath = findBinary(workspaceFolder);
                
                const cmd = `"${binaryPath}" -query "${query}" -fast-ai -output-format=llm-dense -output="${tempFile}"`;
                
                exec(cmd, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`UltraSearch Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    if (fs.existsSync(tempFile)) {
                        vscode.workspace.openTextDocument(tempFile).then(doc => {
                            vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                        });
                    } else {
                        vscode.window.showInformationMessage("Search completed but no output file was generated.");
                    }
                    resolve();
                });
            });
        });
    });

    let quickUrlsDisposable = vscode.commands.registerCommand('ultrasearch.quickUrls', async function () {
        const query = await vscode.window.showInputBox({
            prompt: "Enter search query for Quick 10 URLs (Snippets Only)",
            placeHolder: "e.g., world population 2026"
        });

        if (!query) return;

        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "UltraSearch: Fetching 10 URLs...",
            cancellable: false
        }, async (progress) => {
            return new Promise((resolve, reject) => {
                const tempFile = path.join(os.tmpdir(), `ultra_quick_${Date.now()}.txt`);
                const workspaceFolder = vscode.workspace.workspaceFolders ? vscode.workspace.workspaceFolders[0].uri.fsPath : __dirname;
                const binaryPath = findBinary(workspaceFolder);
                
                const cmd = `"${binaryPath}" -query "${query}" -no-ai -content=false -output-format=llm-dense -output="${tempFile}"`;
                
                exec(cmd, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`UltraSearch Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    if (fs.existsSync(tempFile)) {
                        vscode.workspace.openTextDocument(tempFile).then(doc => {
                            vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                        });
                    } else {
                        vscode.window.showInformationMessage("Search completed but no output file was generated.");
                    }
                    resolve();
                });
            });
        });
    });

    let onlyAIDisposable = vscode.commands.registerCommand('ultrasearch.onlyAI', async function () {
        const query = await vscode.window.showInputBox({
            prompt: "Enter search query for Only AI Overview",
            placeHolder: "e.g., why is sky blue"
        });

        if (!query) return;

        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "UltraSearch: Fetching AI Overview...",
            cancellable: false
        }, async (progress) => {
            return new Promise((resolve, reject) => {
                const tempFile = path.join(os.tmpdir(), `ultra_ai_${Date.now()}.txt`);
                const workspaceFolder = vscode.workspace.workspaceFolders ? vscode.workspace.workspaceFolders[0].uri.fsPath : __dirname;
                const binaryPath = findBinary(workspaceFolder);
                
                const cmd = `"${binaryPath}" -query "${query}" -only-ai -output-format=llm-dense -output="${tempFile}"`;
                
                exec(cmd, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`UltraSearch Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    if (fs.existsSync(tempFile)) {
                        vscode.workspace.openTextDocument(tempFile).then(doc => {
                            vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                        });
                    } else {
                        vscode.window.showInformationMessage("Search completed but no output file was generated.");
                    }
                    resolve();
                });
            });
        });
    });

    context.subscriptions.push(disposable, fastDisposable, quickUrlsDisposable, onlyAIDisposable);
}

function deactivate() {}

module.exports = {
    activate,
    deactivate
}
