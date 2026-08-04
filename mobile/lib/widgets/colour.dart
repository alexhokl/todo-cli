import 'package:flutter/material.dart';

/// Parses a `#RRGGBB` hex string into a [Color].
///
/// Returns `null` when [hex] is empty, malformed, or not exactly six hex
/// digits after an optional leading `#`. The alpha channel is forced to
/// fully opaque (`0xFF`).
Color? parseLabelColour(String hex) {
  var value = hex;
  if (value.startsWith('#')) value = value.substring(1);
  if (value.length != 6) return null;
  final parsed = int.tryParse(value, radix: 16);
  if (parsed == null) return null;
  return Color(0xFF000000 | parsed);
}