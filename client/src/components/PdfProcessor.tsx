import React, { useState } from 'react';
import {
  Button,
  Container,
  Paper,
  Typography,
  CircularProgress,
  TextField,
  RadioGroup,
  FormControlLabel,
  Radio,
  Chip,
  MenuItem,
  Select,
  InputLabel,
  FormControl,
  Box
} from '@mui/material';
import { CloudUpload, PictureAsPdf, Delete } from '@mui/icons-material';
import axios from 'axios';

type PdfOperation = 'merge' | 'split';

const PdfProcessor: React.FC = () => {
  const [operation, setOperation] = useState<PdfOperation>('merge');
  const [mergeFiles, setMergeFiles] = useState<File[]>([]);
  const [splitFile, setSplitFile] = useState<File | null>(null);
  const [splitMode, setSplitMode] = useState<'pages' | 'count'>('pages');
  const [pages, setPages] = useState('');
  const [count, setCount] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Common file upload handler
  const handleFileUpload = (files: FileList | null) => {
    if (!files) return;

    if (operation === 'merge') {
      const newFiles = Array.from(files);
      setMergeFiles(prev => [...prev, ...newFiles]);
    } else {
      setSplitFile(files[0]);
    }
    setError('');
  };

  // Merge handlers
  const handleRemoveMergeFile = (index: number) => {
    setMergeFiles(prev => prev.filter((_, i) => i !== index));
  };

  const handleMerge = async () => {
    if (mergeFiles.length < 2) {
      setError('Please select at least two PDF files to merge');
      return;
    }

    const formData = new FormData();
    mergeFiles.forEach(file => formData.append('pdfs', file));

    try {
      setLoading(true);
      const response = await axios.post('http://localhost:3000/merge-pdfs', formData, {
        responseType: 'blob',
      });

      downloadFile(response.data, 'merged.pdf');
    } catch (err) {
        console.log(err);
        
      setError('Error merging PDFs. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Split handlers
  const handleSplit = async () => {
    if (!splitFile) {
      setError('Please select a PDF file to split');
      return;
    }

    const formData = new FormData();
    formData.append('pdf', splitFile);
    formData.append('mode', splitMode);

    if (splitMode === 'pages') {
      if (!pages) {
        setError('Please enter page ranges');
        return;
      }
      formData.append('pages', pages);
    } else {
      if (!count || isNaN(Number(count))) {
        setError('Please enter a valid page count');
        return;
      }
      formData.append('count', count);
    }

    try {
      setLoading(true);
      const response = await axios.post('http://localhost:3000/split-pdf', formData, {
        responseType: 'blob',
      });

      downloadFile(response.data, 'split_pdfs.zip');
    } catch (err) {
        console.log(err);
        
      setError('Error splitting PDF. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  // Common download handler
  const downloadFile = (data: Blob, filename: string) => {
    const url = window.URL.createObjectURL(new Blob([data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    link.parentNode?.removeChild(link);
    window.URL.revokeObjectURL(url);
  };

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Paper elevation={3} sx={{ p: 4 }}>
        <Typography variant="h4" gutterBottom sx={{ mb: 3 }}>
          PDF Processor
        </Typography>

        {/* Operation Selector */}
        <FormControl fullWidth sx={{ mb: 4 }}>
          <InputLabel>Select Operation</InputLabel>
          <Select
            value={operation}
            label="Select Operation"
            onChange={(e) => setOperation(e.target.value as PdfOperation)}
          >
            <MenuItem value="merge">Merge PDFs</MenuItem>
            <MenuItem value="split">Split PDF</MenuItem>
          </Select>
        </FormControl>

        {/* Merge Interface */}
        {operation === 'merge' && (
          <Box>
            <Button
              variant="contained"
              component="label"
              startIcon={<CloudUpload />}
              sx={{ mb: 2 }}
            >
              Select PDF Files
              <input
                type="file"
                hidden
                multiple
                accept="application/pdf"
                onChange={(e) => handleFileUpload(e.target.files)}
              />
            </Button>

            <Box sx={{ mb: 2 }}>
              {mergeFiles.map((file, index) => (
                <Chip
                  key={index}
                  label={file.name}
                  onDelete={() => handleRemoveMergeFile(index)}
                  deleteIcon={<Delete />}
                  variant="outlined"
                  sx={{ m: 0.5 }}
                />
              ))}
            </Box>

            <Button
              variant="contained"
              color="primary"
              onClick={handleMerge}
              disabled={loading || mergeFiles.length < 2}
              startIcon={loading ? <CircularProgress size={20} /> : <PictureAsPdf />}
            >
              {loading ? 'Merging...' : `Merge ${mergeFiles.length} PDFs`}
            </Button>
          </Box>
        )}

        {/* Split Interface */}
        {operation === 'split' && (
          <Box>
            <Button
              variant="contained"
              component="label"
              startIcon={<CloudUpload />}
              sx={{ mb: 2 }}
            >
              {splitFile ? splitFile.name : 'Select PDF File'}
              <input
                type="file"
                hidden
                accept="application/pdf"
                onChange={(e) => handleFileUpload(e.target.files)}
              />
            </Button>

            <FormControl component="fieldset" fullWidth sx={{ mb: 2 }}>
              <RadioGroup
                value={splitMode}
                onChange={(e) => setSplitMode(e.target.value as 'pages' | 'count')}
              >
                <FormControlLabel 
                  value="pages" 
                  control={<Radio />} 
                  label="Split by Page Ranges (e.g., 1-3,5,7-9)" 
                />
                <FormControlLabel
                  value="count"
                  control={<Radio />}
                  label="Split by Page Count"
                />
              </RadioGroup>
            </FormControl>

            {splitMode === 'pages' ? (
              <TextField
                fullWidth
                label="Page Ranges"
                value={pages}
                onChange={(e) => setPages(e.target.value)}
                sx={{ mb: 2 }}
                placeholder="Example: 1-3,5,7-9"
              />
            ) : (
              <TextField
                fullWidth
                label="Pages per File"
                type="number"
                value={count}
                onChange={(e) => setCount(e.target.value)}
                sx={{ mb: 2 }}
                inputProps={{ min: 1 }}
              />
            )}

            <Button
              variant="contained"
              color="primary"
              onClick={handleSplit}
              disabled={loading || !splitFile}
              startIcon={loading ? <CircularProgress size={20} /> : <PictureAsPdf />}
            >
              {loading ? 'Splitting...' : 'Split PDF'}
            </Button>
          </Box>
        )}

        {/* Error Display */}
        {error && (
          <Typography color="error" sx={{ mt: 2 }}>
            {error}
          </Typography>
        )}
      </Paper>
    </Container>
  );
};

export default PdfProcessor;