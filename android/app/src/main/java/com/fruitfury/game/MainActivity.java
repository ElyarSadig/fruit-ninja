package com.fruitfury.game;

import android.app.Activity;
import android.os.Bundle;
import android.view.View;
import android.view.ViewGroup;

import com.fruitfury.game.mobile.EbitenView;
import com.fruitfury.game.mobile.Mobile;

/** Native fullscreen host for the Ebitengine game. */
public final class MainActivity extends Activity {
    private EbitenView gameView;

    private void enableImmersiveMode() {
        getWindow().getDecorView().setSystemUiVisibility(
            View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION
                | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                | View.SYSTEM_UI_FLAG_FULLSCREEN
                | View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
        );
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        enableImmersiveMode();
        // EbitenView v2.10 sets the Go/Android context itself.
        gameView = new EbitenView(this);
        // Save high scores in /data/user/0/<package>/files, never public storage.
        Mobile.setDataDirectory(getFilesDir().getAbsolutePath());
        setContentView(gameView, new ViewGroup.LayoutParams(
            ViewGroup.LayoutParams.MATCH_PARENT,
            ViewGroup.LayoutParams.MATCH_PARENT
        ));
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            enableImmersiveMode();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (gameView != null) {
            gameView.resumeGame();
        }
    }

    @Override
    protected void onPause() {
        if (gameView != null) {
            gameView.suspendGame();
        }
        super.onPause();
    }
}
