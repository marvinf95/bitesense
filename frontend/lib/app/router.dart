import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/auth_controller.dart';
import '../features/analytics/analytics_screen.dart';
import '../features/auth/login_screen.dart';
import '../features/auth/register_screen.dart';
import '../features/export/export_screen.dart';
import '../features/meals/meal_add_screen.dart';
import '../features/meals/meal_detail_screen.dart';
import '../features/meals/meals_screen.dart';
import '../features/settings/settings_screen.dart';
import '../features/symptoms/symptom_add_screen.dart';
import '../features/symptoms/symptoms_screen.dart';
import 'home_shell.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/meals',
    refreshListenable: ref.watch(authListenableProvider),
    redirect: (context, state) {
      final loggedIn = ref.read(authControllerProvider).isAuthenticated;
      final loggingIn = state.matchedLocation == '/login' ||
          state.matchedLocation == '/register';
      if (!loggedIn && !loggingIn) return '/login';
      if (loggedIn && loggingIn) return '/meals';
      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
      GoRoute(path: '/register', builder: (_, __) => const RegisterScreen()),
      ShellRoute(
        builder: (context, state, child) => HomeShell(child: child),
        routes: [
          GoRoute(
            path: '/meals',
            builder: (_, __) => const MealsScreen(),
            routes: [
              GoRoute(path: 'add', builder: (_, __) => const MealAddScreen()),
              GoRoute(
                path: ':id',
                builder: (_, state) =>
                    MealDetailScreen(id: state.pathParameters['id']!),
              ),
            ],
          ),
          GoRoute(
            path: '/symptoms',
            builder: (_, __) => const SymptomsScreen(),
            routes: [
              GoRoute(
                path: 'add',
                builder: (_, __) => const SymptomAddScreen(),
              ),
            ],
          ),
          GoRoute(
            path: '/analytics',
            builder: (_, __) => const AnalyticsScreen(),
          ),
          GoRoute(path: '/export', builder: (_, __) => const ExportScreen()),
          GoRoute(
            path: '/settings',
            builder: (_, __) => const SettingsScreen(),
          ),
        ],
      ),
    ],
  );
});
