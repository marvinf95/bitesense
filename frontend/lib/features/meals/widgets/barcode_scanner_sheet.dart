import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

class BarcodeScannerSheet extends StatefulWidget {
  const BarcodeScannerSheet({super.key});
  @override
  State<BarcodeScannerSheet> createState() => _BarcodeScannerSheetState();
}

class _BarcodeScannerSheetState extends State<BarcodeScannerSheet> {
  final _controller = MobileScannerController();
  bool _handled = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: MediaQuery.of(context).size.height * 0.7,
      child: MobileScanner(
        controller: _controller,
        onDetect: (capture) {
          if (_handled) return;
          for (final code in capture.barcodes) {
            final raw = code.rawValue;
            if (raw == null || raw.isEmpty) continue;
            _handled = true;
            Navigator.of(context).pop(raw);
            return;
          }
        },
      ),
    );
  }
}
