"use client"

import { Box, Button } from '@mui/material';

const ExportCSV = () => {
    const handleExport = async () => {
        try {
            const apiUrl = process.env.NEXT_PUBLIC_BACK_ORIGIN
            const response = await fetch(`${apiUrl}/download-csv`);
            if (response.ok) {
                const blob = await response.blob();
                const contentDisposition = response.headers.get('Content-Disposition');
                let fileName = 'Anonymized-ID.csv'; // default failename

                if (contentDisposition && contentDisposition.includes('filename=')) {
                    const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(contentDisposition);
                    if (matches != null && matches[1]) { 
                      fileName = matches[1].replace(/['"]/g, '');
                    }
                  }

                const url = window.URL.createObjectURL(new Blob([blob]));
                const a = document.createElement('a');
                a.href = url;
                a.download = fileName;
                document.body.appendChild(a);
                a.click();
                window.URL.revokeObjectURL(url);
                a.remove();
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
                患者IDと匿名化IDの対応表をダウンロード
            </Button>
        </Box>
    );
}

export default ExportCSV;
