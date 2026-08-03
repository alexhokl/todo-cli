import 'package:flutter/material.dart';

import 'package:todo/proto/item.pb.dart';

/// Returns a status [Icon] reflecting [item]'s triage state:
/// done → check_circle_outline, triaged (has priority) → low_priority,
/// untriaged → radio_button_unchecked. Decorative; callers should supply
/// the accessible label via the surrounding widget's title.
Icon statusIconFor(Item item) {
  if (item.done) {
    return const Icon(Icons.check_circle_outline);
  }
  if (item.hasPriority()) {
    return const Icon(Icons.low_priority);
  }
  return const Icon(Icons.radio_button_unchecked);
}