'use client';

import React, { useState, ChangeEvent, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';
import { 
  Button,
  TextField,
  Typography,  
  Paper,
  Stack,  
} from '@mui/material';
import { uploadFiles } from '@/lib/uploadFiles';
//import JSZip from "jszip";

type FormValuesType = {
  password: string;
  passwordConfirmation: string;
};

const Form = () => {
  // フォーム用のフックを利用して、各種メソッドと状態を取得
  const { handleSubmit: onSubmit, watch, formState: { errors }, register } = useForm<FormValuesType>();
  const passwordValue = watch('password', '');

  const [files, setFiles] = useState<File[]>([]);
  const [uploading, setUploading] = useState(false);  
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);

  // ファイル変更時のハンドラー
  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFiles(Array.from(e.target.files));
    }
  };

  // フォーム送信時のハンドラー
  const handleFormSubmit = async (data: FormValuesType) => {
    if (files.length === 0) return;
    setUploading(true);
    await uploadFiles(data.password, data.passwordConfirmation, files);
    setUploading(false);
  };

  return (
    <Paper elevation={3} sx={{ maxWidth: '600px', margin: 'auto', padding: 4, marginTop: 8 }}>
      <form noValidate onSubmit={onSubmit(handleFormSubmit)}>
        <Stack spacing={2}> 
          <TextField
            label="パスワード"
            type='password'
            variant="filled"
            fullWidth
            error={!!errors.password}
            helperText={errors.password?.message}
            {...register('password', { required: 'パスワードを入力してください' })}
          />

          <TextField
            label="パスワードの確認"
            type='password'
            variant="filled"
            fullWidth
            error={!!errors.passwordConfirmation}
            helperText={errors.passwordConfirmation?.message}
            {...register('passwordConfirmation', {
              required: 'パスワードの確認を入力してください',
              validate: value =>
                value === passwordValue || 'パスワードが一致しません'
            })}
          />        
        </Stack>            

        <input
          type="file"
          multiple
          accept=".mwf,.MWF"
          onChange={handleFileChange}
          ref={fileInputRef}
          style={{ display: 'none' }}
          /* @ts-expect-error */
          directory=""
          webkitdirectory=""
          id="upload-form"
        />
        <label htmlFor="upload-form">
          <Button variant="outlined" component="span" sx={{ padding: '10px 20px', borderColor: '#1976d2', color: '#1976d2' }}>
            フォルダを選択
          </Button>
        </label>        
        <Button 
          onClick={onSubmit(handleFormSubmit)}
          disabled={files.length === 0 || uploading} 
          variant="contained" 
          sx={{ 
            padding: '10px 20px', 
            backgroundColor: uploading ? '#ccc' : '#1976d2', 
            color: '#fff', 
            borderRadius: 1, 
            '&:hover': { backgroundColor: '#1565c0' }, 
            '&:disabled': { backgroundColor: '#ccc', color: '#fff' }
          }}
        >
          {uploading ? 'Uploading...' : 'アップロード'}
        </Button>
        {files.length > 0 && (
          <>
            <Typography variant="body1" sx={{ mt: 2 }}>
              合計ファイル数: {files.length}
            </Typography>

            <Typography variant="body1" sx={{ mt: 1 }}>
              拡張子ごとのファイル数:
            </Typography>
            <ul>
              {Object.entries(files.reduce((acc, file) => {
                const ext = file.name.split('.').pop() || 'unknown';
                acc[ext] = (acc[ext] || 0) + 1;
                return acc;
              }, {} as Record<string, number>)).map(([ext, count]) => (
                <li key={ext}>{ext}: {count}</li>
              ))}
            </ul>
          </>
        )}
      </form>
    </Paper>
  );
};

export default Form;
