package util

import hudson.AbortException

class PipelineWhenException extends AbortException{
    public PipelineWhenException(String message)
    {
        super(message);
    }
}
