import React from 'react';
import { Link } from 'react-router-dom';
import { AppBar, Toolbar, Typography, Button } from '@mui/material';

const Navigation: React.FC = () => {
    return (
        <AppBar position="static">
            <Toolbar>
                <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
                    File Processor
                </Typography>
                <Button color="inherit" component={Link} to="/image-processor">
                    Image Processor
                </Button>
                <Button color="inherit" component={Link} to="/pdf-processor">
                    PDF Processor
                </Button>
            </Toolbar>
        </AppBar>
    );
};

export default Navigation;