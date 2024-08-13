"use client"

import { Box, Button } from '@mui/material';

const ExportCSV = () => {
    const handleExport = async () => {
        try {
            const response = await fetch('http://localhost:8080/download-csv');
            if (response.ok) {
                // レスポンスを処理する（例: CSVファイルのダウンロード）
                const blob = await response.blob();
                const url = window.URL.createObjectURL(new Blob([blob]));
                const link = document.createElement('a');
                link.href = url;
                link.setAttribute('download', 'Anonymized-ID.csv'); // ダウンロードファイル名
                document.body.appendChild(link);
                link.click();
                //link.parentNode.removeChild(link);
            } else {
                console.error('Failed to export CSV');
            }
        } catch (error) {
            console.error('Error:', error);
        }
    };

    return (
        <Box sx={{ mt: 4, textAlign: 'center' }}>
            <Button variant="contained" color="primary" onClick={handleExport}>
                IDと匿名化IDの対応表をダウンロード
            </Button>
        </Box>
    );
}

export default ExportCSV;
