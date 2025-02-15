import React, { useState } from 'react';
import axios from 'axios';
import { Button, Container, Paper, Typography, TextField, InputLabel, FormControl, Select, MenuItem, CircularProgress } from '@mui/material';
import { CloudUpload, Image } from '@mui/icons-material';

type Operation = 'compress' | 'resize' | 'jpg-to-png' | 'png-to-jpg'  | 'image-to-pdf' | 'transparent-background';

const ImageProcessor: React.FC = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [operation, setOperation] = useState<Operation>('compress');
  const [quality, setQuality] = useState(50);
  const [width, setWidth] = useState('');
  const [height, setHeight] = useState('');
  const [resultImage, setResultImage] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files?.[0]) {
      setSelectedFile(e.target.files[0]);
      setError('');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedFile) {
      setError('Please select a file');
      return;
    }

    const formData = new FormData();
    formData.append('image', selectedFile);

    let url = 'http://localhost:3000/';
    switch (operation) {
      case 'compress':
        url += 'compress';
        formData.append('quality', quality.toString());
        break;
      case 'resize':
        url += 'resize';
        formData.append('width', width);
        formData.append('height', height);
        break;
      case 'jpg-to-png':
        url += 'convert/jpg-to-png';
        break;
      case 'png-to-jpg':
        url += 'convert/png-to-jpg';
        break;
      case 'image-to-pdf':
        url += 'convert/to-pdf';
        break;
      case 'transparent-background':
          url += 'transparent';
          break;
    }

    try {
      setLoading(true);
      const response = await axios.post(url, formData, {
          responseType: 'blob',
      });

      if (operation === 'image-to-pdf') {
           // Handle PDF download
           const url = window.URL.createObjectURL(new Blob([response.data]));
           const link = document.createElement('a');
           link.href = url;
           link.setAttribute('download', 'converted.pdf');
           document.body.appendChild(link);
           link.click();
           link.parentNode?.removeChild(link);
           window.URL.revokeObjectURL(url);
      } else {
          const imageUrl = URL.createObjectURL(response.data);
          setResultImage(imageUrl);
      }
      setError('');
    } catch (err) {
        console.log(err);
        
      setError('Error processing image. Please try again.');
      setResultImage(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Paper elevation={3} sx={{ p: 4 }}>
        <Typography variant="h4" gutterBottom>
          Image Processor
        </Typography>

        <form onSubmit={handleSubmit}>
          <FormControl fullWidth margin="normal">
            <Button
              component="label"
              variant="contained"
              startIcon={<CloudUpload />}
            >
              Upload Image
              <input
                type="file"
                hidden
                accept="image/jpeg, image/png"
                onChange={handleFileChange}
              />
            </Button>
            {selectedFile && (
              <Typography variant="body2" sx={{ mt: 1 }}>
                Selected: {selectedFile.name}
              </Typography>
            )}
          </FormControl>

          <FormControl fullWidth margin="normal">
            <InputLabel>Operation</InputLabel>
            <Select
              value={operation}
              label="Operation"
              onChange={(e) => setOperation(e.target.value as Operation)}
            >
              <MenuItem value="compress">Compress</MenuItem>
              <MenuItem value="resize">Resize</MenuItem>
              <MenuItem value="jpg-to-png">Convert JPG to PNG</MenuItem>
              <MenuItem value="png-to-jpg">Convert PNG to JPG</MenuItem>
              <MenuItem value="image-to-pdf">Convert to PDF</MenuItem>
              <MenuItem value="transparent-background">Make Background Transparent</MenuItem>
            </Select>
          </FormControl>

          {operation === 'compress' && (
            <TextField
              fullWidth
              margin="normal"
              label="Quality (1-100)"
              type="number"
              value={quality}
              onChange={(e) => {
                const value = Math.min(Math.max(Number(e.target.value), 1), 100);
                setQuality(value);
              }}
              inputProps={{ min: 1, max: 100 }}
            />
          )}

          {operation === 'resize' && (
            <div style={{ display: 'flex', gap: '1rem' }}>
              <TextField
                fullWidth
                margin="normal"
                label="Width"
                type="number"
                value={width}
                onChange={(e) => setWidth(e.target.value)}
              />
              <TextField
                fullWidth
                margin="normal"
                label="Height"
                type="number"
                value={height}
                onChange={(e) => setHeight(e.target.value)}
              />
            </div>
          )}

          {error && (
            <Typography color="error" sx={{ mt: 2 }}>
              {error}
            </Typography>
          )}

          <Button
            type="submit"
            variant="contained"
            color="primary"
            disabled={loading || !selectedFile}
            sx={{ mt: 3 }}
            startIcon={loading ? <CircularProgress size={20} /> : <Image />}
          >
            {loading ? 'Processing...' : 'Process Image'}
          </Button>
        </form>

        {resultImage && (
          <div style={{ marginTop: '2rem' }}>
            <Typography variant="h6" gutterBottom>
              Processed Image
            </Typography>
            <a href={resultImage} download>
              <img
                src={resultImage}
                alt="Processed result"
                style={{ maxWidth: '100%', maxHeight: '400px' }}
              />
            </a>
          </div>
        )}
      </Paper>
    </Container>
  );
};

export default ImageProcessor;