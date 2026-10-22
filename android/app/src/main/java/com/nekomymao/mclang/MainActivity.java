package com.nekomymao.mclang;
import android.os.Bundle;
import android.webkit.WebSettings;
import android.webkit.WebView;
import androidx.appcompat.app.AppCompatActivity;
public class MainActivity extends AppCompatActivity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        WebView w = new WebView(this);
        w.getSettings().setJavaScriptEnabled(true);
        w.loadUrl("http://10.0.2.2:9876");
        setContentView(w);
    }
}
