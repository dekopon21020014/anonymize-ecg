'use client';

//import React, { useState } from 'react';
import { Container, Box, Typography, Button } from '@mui/material';
import Form from '@/components/Form';
import ExoprtCSV from '@/components/ExportCSV';

const TopPage = () => {    
    return (
        <Container maxWidth="md">
            <Box component="main" sx={{ my: 4, maxWidth: '600px', mx: 'auto' }}>
                <Typography 
                    variant="h4" 
                    component="h1" 
                    gutterBottom 
                    sx={{ 
                        fontWeight: 'bold',
                        textAlign: 'center'  // ここを追加
                    }}
                >
                    操作手順
                </Typography>
                <Typography variant="h5"> 1.パスワードを入力してください．</Typography>
                <Typography variant="h5"> 2.データの含まれるフォルダを選択してください．</Typography>
                関係ないファイルやフォルダが含まれていても構いません．<br  />それらは匿名化処理の際に除外されます<div className=""></div>
                <Typography variant="h5"> 3.アップロードしてください．　</Typography>
                <li>処理が終了するとzipファイルが作成されます．</li>
                <Typography variant="h5">4.zipファイルをUSBに保存してください</Typography>     
            </Box>
            <Form/>{/* Formは自作のコンポーネント*/}
            <ExoprtCSV/>
        </Container>
    );
};

export default TopPage;
