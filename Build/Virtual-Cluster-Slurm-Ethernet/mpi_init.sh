module load mpi

# Change MPI parameter for ethernet tcp only
OMPI_MCA_btl=tcp,self
export OMPI_MCA_btl
