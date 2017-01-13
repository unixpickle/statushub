package com.aqnichol.statushub;

import android.app.AlertDialog;
import android.content.DialogInterface;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.wearable.PutDataMapRequest;
import com.google.android.gms.wearable.PutDataRequest;
import com.google.android.gms.wearable.Wearable;

public class MainActivity extends AppCompatActivity {

    private final static String TAG = "StatusHub";

    private EditText rootURL;
    private EditText password;
    private GoogleApiClient client;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        final Button b = (Button)findViewById(R.id.saveButton);
        b.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                save();
            }
        });
        password = (EditText)findViewById(R.id.password);
        rootURL = (EditText)findViewById(R.id.rootURL);
        client = new GoogleApiClient.Builder(this)
                .addConnectionCallbacks(new GoogleApiClient.ConnectionCallbacks() {
                    @Override
                    public void onConnected(Bundle connectionHint) {
                        Log.d(TAG, "onConnected: " + connectionHint);
                        // Now you can use the Data Layer API
                    }

                    @Override
                    public void onConnectionSuspended(int cause) {
                        Log.d(TAG, "onConnectionSuspended: " + cause);
                    }
                })
                .addOnConnectionFailedListener(new GoogleApiClient.OnConnectionFailedListener() {
                    @Override
                    public void onConnectionFailed(ConnectionResult result) {
                        Log.d(TAG, "onConnectionFailed: " + result);
                    }
                })
                .addApi(Wearable.API)
                .build();
        client.connect();
    }

    private void save() {
        if (!client.isConnected()) {
            errorDialog(R.string.ps_disconn_error);
            return;
        }
        PutDataMapRequest req = PutDataMapRequest.create("/shhost");
        req.getDataMap().putString("rootURL", rootURL.getText().toString());
        req.getDataMap().putString("password", password.getText().toString());
        PutDataRequest preq = req.asPutDataRequest();
        Wearable.DataApi.putDataItem(client, preq);
    }

    private void errorDialog(int msgResource) {
        AlertDialog.Builder builder = new AlertDialog.Builder(this);
        builder.setMessage(msgResource);
        builder.setPositiveButton(R.string.ok, new DialogInterface.OnClickListener() {
            @Override
            public void onClick(DialogInterface dialog, int which) {
                dialog.dismiss();
            }
        });
        AlertDialog d = builder.create();
        d.show();
    }
}
