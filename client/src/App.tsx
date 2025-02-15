import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import ImageProcessor from './components/ImageProcessor';
import PdfProcessing from './components/PdfProcessor';
import Navigation from './components/Navigation'; // Optional navigation component

function App() {
    return (
        <Router>
            <Navigation />
            <Routes>
                <Route path="/" element={<ImageProcessor />} />
                <Route path="/image-processor" element={<ImageProcessor />} />
                <Route path="/pdf-processor" element={<PdfProcessing />} />
            </Routes>
        </Router>
    );
}

export default App;