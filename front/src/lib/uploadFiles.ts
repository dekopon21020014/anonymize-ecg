import JSZip from "jszip";

export async function uploadFiles(
    password: string,
    passwordConfirmation: string,
    files: File[]
) {
    return new Promise<void>((resolve, reject) => {
        const ws = new WebSocket("ws://localhost:8080/upload");

        ws.onopen = () => {
            // 認証情報を送信
            ws.send(
                JSON.stringify({
                    type: "credentials",
                    password,
                    passwordConfirmation,
                })
            );
        };

        ws.onmessage = async (event) => {
            if (event.data === "ok") {
                ws.onmessage = downloadZip(ws)
                // ファイル送信を開始
                const chunkSize = 1000;
                let startIndex = 0;

                while (startIndex < files.length) {
                    const endIndex = Math.min(startIndex + chunkSize, files.length);
                    const fileChunk = files.slice(startIndex, endIndex);

                    const zip = new JSZip();

                    // 各ファイルを ZIP に追加
                    for (const file of fileChunk) {
                        const fileBuffer = await file.arrayBuffer();
                        zip.file(file.name, fileBuffer);
                    }

                    // ZIP ファイルを生成
                    const zipBlob = await zip.generateAsync({ type: "blob" });

                    // ZIP ファイルを送信
                    ws.send(zipBlob);

                    // 次の 1000 ファイルに進む
                    startIndex = endIndex;
                }

                // 全てのファイルを送信した後に終了メッセージを送信
                ws.send("end");
            } else {
                console.error("WebSocket Error:", event.data);
                reject(new Error("WebSocket Error"));
            }
        };

        ws.onerror = (error) => {
            console.error("WebSocket Error:", error);
            reject(error);
        };

        ws.onclose = () => {
            console.log("WebSocket connection closed.");
            resolve();
        };
    });
}


export function downloadZip(ws: WebSocket) {
    return async (event: MessageEvent) => {
    const message = event.data;
  
    if (typeof message === 'string') {
        try {
            const metaData = JSON.parse(message);
            if (metaData.fileName && metaData.fileType) {
            console.log(`Receiving file: ${metaData.fileName}`);
            // メタデータを受信した場合、次のバイナリメッセージを待つ
            ws.onmessage = (event) => {
                if (typeof event.data === 'object') {
                    // 受信したデータが ZIP ファイルなら、それを Blob として保存
                    const blob = new Blob([event.data], { type: metaData.fileType });
                    const url = URL.createObjectURL(blob);
        
                    // ダウンロードリンクを作成してクリック
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = metaData.fileName; // メタデータからファイル名を取得
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
        
                    // Blob URL の解放
                    URL.revokeObjectURL(url);
                }
            };
            } else {
                console.log('Authentication successful');
            }
        } catch (e) {
            console.log('Error parsing message:', e);
        }
    }
  };
}