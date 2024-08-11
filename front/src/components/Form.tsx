'use client';

import React, { useState, ChangeEvent, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';
import { 
  Button,
  Box,
  TextField,
  Typography,  
  Paper,
  Stack,  
} from '@mui/material';

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

    const formData = new FormData();
    files.forEach((file, index) => {
      formData.append(`file${index}`, file);
    });    

    formData.append('password', data.password);
    formData.append('passwordConfirmation', data.passwordConfirmation);

    try {
      const response = await fetch('http://localhost:8080/', {
        method: 'POST',
        body: formData,
      });

      if (response.ok) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'anonymized-files.zip';
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        a.remove();

        console.log('Files downloaded successfully');
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
        setFiles([]);
        router.push('/');
      } else {
        console.error('Upload failed');
      }
    } catch (error) {
      console.error('Error during upload:', error);
    } finally {
      setUploading(false);
    }
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
        {files.length > 0 && (
          <Typography variant="body1" sx={{ mt: 2 }}>
            Selected files: {files.map(file => file.name).join(', ')}
          </Typography>
        )}
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
      </form>
    </Paper>
  );
};

export default Form;


/*
'use client';

import React, { useState, ChangeEvent, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { useRouter } from 'next/navigation';
import { 
  Button,
  Box,
  TextField,
  Typography,  
  Paper,
  Stack,  
} from '@mui/material';

type FormValuesType = {
  password: string;
  passwordConfirmation: string;
};

export const Form = () => {
  // ここから
  const {handleSubmit: onSubmit, watch, formState: { errors }, register } = useForm<FormValuesType>();
  const passwordValue = watch('password', '');

  // ここまで
  const [files, setFiles] = useState<File[]>([]);
  const [uploading, setUploading] = useState(false);  
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      setFiles(Array.from(e.target.files));
    }
  };

  const handleSubmit = async () => {
    if (files.length === 0) return;
    setUploading(true);

    const formData = new FormData();
    files.forEach((file, index) => {
      formData.append(`file${index}`, file);
    });

    try {
      const response = await fetch('http://localhost:8080/', {
        method: 'POST',
        body: formData,
      });

      if (response.ok) {
        // ZIPファイルのバイナリデータを取得
        const blob = await response.blob();
    
        // BlobデータをURLに変換
        const url = window.URL.createObjectURL(blob);
    
        // ダウンロード用のリンクを作成
        const a = document.createElement('a');
        a.href = url;
        a.download = 'anonymized-files.zip'; // 任意のファイル名を設定
        document.body.appendChild(a);
        a.click();
    
        // ダウンロード後にクリーンアップ
        window.URL.revokeObjectURL(url);
        a.remove();
    
        console.log('Files downloaded successfully');
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
        setFiles([]); // Reset the files state
        router.push('/'); // Redirect after successful upload
      } else {
        console.error('Upload failed');
      }
    } catch (error) {
      console.error('Error during upload:', error);
    } finally {
      setUploading(false);
    }
  };

  return (
    //<Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
      <Paper elevation={3} sx={{ maxWidth: '600px', margin: 'auto', padding: 4, marginTop: 8 }}>
            <form noValidate onSubmit={onSubmit(handleSubmit)}>
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
                   
                    <Box textAlign='right'>
                        <Button
                            variant="contained"
                            type='submit'
                            //disabled={props.isLoading}
                        >
                            送信
                        </Button>
                    </Box>
                </Stack>            

      <input
        type="file"
        multiple
        accept=".mwf,.MWF"
        onChange={handleFileChange}
        ref={fileInputRef}
        style={{ display: 'none' }}
        /* @ts-expect-error 
        directory=""
        webkitdirectory=""
        id="upload-form"
      />
      <label htmlFor="upload-form">
        <Button variant="outlined" component="span" sx={{ padding: '10px 20px', borderColor: '#1976d2', color: '#1976d2' }}>
          Choose Files
        </Button>
      </label>
      {files.length > 0 && (
        <Typography variant="body1" sx={{ mt: 2 }}>
          Selected files: {files.map(file => file.name).join(', ')}
        </Typography>
      )}
      <Button 
        //onClick={handleSubmit} 
        onClick={onSubmit(handleSubmit)}
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
        {uploading ? 'Uploading...' : 'Submit'}
      </Button>
      </form>
      </Paper>
    // </Box>
  );
};

export default Form;
*/