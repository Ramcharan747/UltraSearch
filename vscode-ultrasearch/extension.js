const vscode = require('vscode');
const { exec } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

function activate(context) {
    let disposable = vscode.commands.registerCommand('ultrasearch.search', async function () {
        const query = await vscode.window.showInputBox({
            prompt: "Enter search query for UltraSearch",
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
                
                // Assuming ultrasearch binary is available in the workspace or system path
                const workspaceFolder = vscode.workspace.workspaceFolders ? vscode.workspace.workspaceFolders[0].uri.fsPath : __dirname;
                // Go up one level from vscode-ultrasearch to find the binary
                const binaryPath = path.join(workspaceFolder, '..', 'ultrasearch');
                
                const cmd = `"${binaryPath}" -query "${query}" -output-format=llm-dense -output="${tempFile}"`;
                
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
            prompt: "Enter search query for Fast AI Overview",
            placeHolder: "e.g., speed of light in vacuum"
        });

        if (!query) return;

        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "UltraSearch: Fetching Fast AI Overview...",
            cancellable: false
        }, async (progress) => {
            return new Promise((resolve, reject) => {
                const tempFile = path.join(os.tmpdir(), `ultra_fast_${Date.now()}.txt`);
                const workspaceFolder = vscode.workspace.workspaceFolders ? vscode.workspace.workspaceFolders[0].uri.fsPath : __dirname;
                const binaryPath = path.join(workspaceFolder, '..', 'ultrasearch');
                
                // Note the -fast-ai flag which triggers the new fast mode
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

    context.subscriptions.push(disposable, fastDisposable);
}

function deactivate() {}

module.exports = {
    activate,
    deactivate
}
